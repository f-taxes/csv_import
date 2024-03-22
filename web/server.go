package web

import (
	"context"
	"embed"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"gitea.codeblob.work/pk/gut/conf"
	iu "gitea.codeblob.work/pk/gut/irisutils"
	"github.com/f-taxes/csv_import/global"
	"github.com/f-taxes/csv_import/grpc_client"
	"github.com/f-taxes/csv_import/processors/ftx"
	"github.com/f-taxes/csv_import/proto"
	"github.com/kataras/golog"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/view"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var schemas = []global.SchemaProcessor{
	{Name: "ftx_deposits", Label: "FTX Deposit History"},
	{Name: "ftx_withdrawals", Label: "FTX Withdrawal History"},
	{Name: "ftx_transactions", Label: "FTX Transaction History", Processor: ftx.NewTransactionProcessor()},
}

func Start(address string, webAssets embed.FS) {
	if conf.App.Bool("debug") {
		global.SetGoLogDebugFormat()
		golog.SetLevel("debug")
		golog.Info("Debug logging is enabled!")
	}

	app := iris.New()
	app.Use(iris.Compression)
	app.SetRoutesNoLog(true)

	registerFrontend(app, webAssets)

	app.Get("/schemes", func(ctx iris.Context) {
		ctx.JSON(iu.Resp{
			Result: true,
			Data:   schemas,
		})
	})

	app.Post("/upload", func(ctx iris.Context) {
		storageRoot := "./tmp_uploads"
		os.MkdirAll(storageRoot, 0700)

		files := []global.ImportFile{}

		account := ctx.FormValue("account")
		schema := ctx.FormValue("schema")

		idx := 0
		_, _, err := ctx.UploadFormFiles(storageRoot, func(ctx iris.Context, file *multipart.FileHeader) bool {
			uniqueName := primitive.NewObjectID().Hex()

			files = append(files, global.ImportFile{
				File:     filepath.Join(storageRoot, uniqueName),
				RealName: file.Filename,
				Account:  account,
				Schema:   schema,
			})

			file.Filename = uniqueName
			idx++
			return true
		})

		if err != nil {
			golog.Errorf("Failed to upload file(s): %v", err)

			ctx.JSON(iu.Resp{
				Result: false,
				Data:   []string{fmt.Sprintf("Upload failed: %v", err.Error())},
			})
			return
		}

		go func() {
			for _, file := range files {
				for i := range schemas {
					if schemas[i].Name == file.Schema {
						contents, err := os.ReadFile(file.File)
						if err != nil {
							golog.Errorf("Failed to read imported file %s: %v", file.File, err)

							ctx.JSON(iu.Resp{
								Result: false,
								Data:   []string{fmt.Sprintf("Upload failed: %v", err.Error())},
							})
							return
						}

						jobId := primitive.NewObjectID().Hex()
						grpc_client.GrpcClient.ShowJobProgress(context.Background(), &proto.JobProgress{
							ID:       jobId,
							Label:    fmt.Sprintf("Importing %s", file.RealName),
							Progress: "-1",
						})

						schemas[i].Processor.Parse(contents, file.Account, file.RealName)

						grpc_client.GrpcClient.ShowJobProgress(context.Background(), &proto.JobProgress{
							ID:       jobId,
							Progress: "100",
						})
					}
				}
			}
		}()

		ctx.JSON(iu.Resp{
			Result: true,
		})
	})

	if err := app.Listen(address); err != nil {
		golog.Fatal(err)
	}
}

func registerFrontend(app *iris.Application, webAssets embed.FS) {
	var frontendTpl *view.HTMLEngine
	useEmbedded := conf.App.Bool("embedded")

	if useEmbedded {
		golog.Debug("Using embedded web sources")
		embeddedFs := iris.PrefixDir("frontend-dist", http.FS(webAssets))
		frontendTpl = iris.HTML(embeddedFs, ".html")
		app.HandleDir("/assets", embeddedFs)
	} else {
		golog.Debug("Using external web sources")
		frontendTpl = iris.HTML("./frontend-dist", ".html")
		app.HandleDir("/assets", "frontend-dist")
	}

	golog.Debug("Automatic reload of web sources is enabled")
	frontendTpl.Reload(conf.App.Bool("debug"))
	app.RegisterView(frontendTpl)
	app.OnAnyErrorCode(index)

	app.Get("/", index)
	app.Get("/{p:path}", index)
}

func index(ctx iris.Context) {
	ctx.View("index.html")
}
