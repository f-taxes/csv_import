/**
@license
Copyright (c) 2024 trading_peter
This program is available under Apache License Version 2.0
*/

import '@tp/tp-icon/tp-icon.js';
import { LitElement, html, css } from 'lit';
import { upload } from '@tp/helpers/upload-files.js';
import icons from '../icons.js';
import shared from '../styles/shared.js';
import { Deferred } from '../helpers/deferred.js';

class TpUploadItem extends upload(LitElement) {
  static get styles() {
    return [
      shared,
      css`
        :host {
          display: block;
        }

        :host([uploading]) {
          pointer-events: none;
          opacity: 0.7;
        }

        .name {
          flex: 1;
        }

        .file {
          display: flex;
          align-items: center;
          padding: 10px;
          background-color: var(--card-box-background);
          border-radius: 10px;
        }

        .file tp-input,
        .file tp-dropdown {
          margin-bottom: 0;
        }

        .file tp-input {
          margin-right: 20px;
        }

        .file > div {
          display: flex;
          flex-direction: row;
          align-items: center;
        }

        .icon {
          margin-right: 10px;
        }

        .remove {
          margin-left: 20px;
        }
      `
    ];
  }

  render() {
    const { file, percent, schemes } = this;

    return html`
      <tp-form class="pending" @submit=${e => this.startUpload(e, file)}>
        <form class="file">
          <div class="icon">
            <tp-icon .icon=${icons.file}></tp-icon>
            <input type="hidden" name="">
          </div>
          <div class="name">
            ${file.name} ${percent ? html`(${percent}%)` : null}
          </div>
          <div>
            <tp-input name="account" required errorMessage="required"}>
              <input type="text" placeholder="Account Label">
            </tp-input>
          </div>
          <div>
            <tp-dropdown name="schema" .default=${schemes[0].name} .items=${schemes.map(s => ({ value: s.name, label: s.label }))}></tp-dropdown>
          </div>
          <div class="remove">
            <tp-icon .icon=${icons.close} @click=${this.removeFile}></tp-icon>
          </div>
        </form>
      </tp-form>
    `;
  }

  static get properties() {
    return {
      file: { type: Object },
      percent: { type: Number },
      uploading: { type: Boolean, reflect: true },
      schemes: { type: Array },
    };
  }

  constructor() {
    super();
    this._boundProgress = this.onProgress.bind(this);
    this._boundFinished = this.onFinished.bind(this);
  }

  connectedCallback() {
    super.connectedCallback();
    this.addEventListener('upload-progress', this._boundProgress);
    this.addEventListener('upload-finished', this._boundFinished);
  }

  disconnectedCallback() {
    super.disconnectedCallback();
    this.removeEventListener('upload-progress', this._boundProgress);
    this.removeEventListener('upload-finished', this._boundFinished);
  }

  onProgress(e) {
    this.percent = e.detail.percent;
  }

  onFinished() {
    this.percent = '100';
  }

  upload() {
    return new Promise(async (resolve, reject) => {
      this.d = new Deferred();
      this.shadowRoot.querySelector('tp-form').submit();
      await this.d.promise;
      resolve();
    });
  }

  /**
   * Start upload of a single file.
   * 
   * @param {Event} e 
   * @param {File} file 
   */
  async startUpload(e, file) {
    this.uploading = true;
    await this.uploadFiles('/upload', [ file ], e.detail);
    this.d.resolve();
  }

  removeFile() {
    this.dispatchEvent(new CustomEvent('remove', { detail: this.file, bubbles: true, composed: true }));
  }
}

window.customElements.define('tp-upload-item', TpUploadItem);