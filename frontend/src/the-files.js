/**
@license
Copyright (c) 2022 trading_peter
This program is available under Apache License Version 2.0
*/

import '@tp/tp-button/tp-button.js';
import '@tp/tp-form/tp-form.js';
import '@tp/tp-dropdown/tp-dropdown.js';
import '@tp/tp-input/tp-input.js';
import '@tp/tp-icon/tp-icon.js';
import './elements/tp-upload-item.js';
import shared from './styles/shared.js';
import { LitElement, html, css } from 'lit';
import { fetchMixin } from '@tp/helpers/fetch-mixin.js';
import { DomQuery } from '@tp/helpers/dom-query.js';

class TheFiles extends fetchMixin(DomQuery(LitElement)) {
  static get styles() {
    return [
      shared,
      css`
        :host {
          display: flex;
          flex-direction: column;
          padding: 20px;
          flex: 1;
        }

        .upload {
          display: flex;
          flex-direction: column;
          margin-top: 20px;
          flex: 1;
        }

        #uploadBtn:not([visible]) {
          display: none;
        }

        tp-upload-item + tp-upload-item {
          margin-top: 10px;
        }

        .icon {
          display: flex;
          align-items: center;
          width: 24px;
          margin-left: 20px;
        }

        .icon.smaller {
          --tp-icon-width: 18px;
          --tp-icon-height: 18px;
        }

        .name {
          flex: 1;
          margin-right: 10px;
        }

        #fakeUpload {
          display: none;
        }
      `
    ];
  }

  render() {
    const { mimeTypes, schemes } = this;

    const files = this.files || [];

    return html`
      <h2>Import a CSV file</h2>
      <div class="upload" @remove=${this.removeFile}>
        ${files.map(file => html`
          <tp-upload-item .file=${file} .schemes=${schemes}></tp-upload-item>
        `)}
      </div>

      <div class="buttons-justified">
        <tp-button @click=${this.browse}>Add Files</tp-button>
        <tp-button @click=${this.reset}>Reset</tp-button>
        <tp-button @click=${this.startUpload}>Upload</tp-button>
      </div>

      <input type="file" multiple accept=${mimeTypes} id="fakeUpload" @change=${() => this.filesSelected()}>
    `;
  }

  static get properties() {
    return {
      schemes: { type: Array },
      files: { type: Array },
      mimeTypes: { type: Array },
    };
  }

  constructor() {
    super();
    this.mimeTypes = ['text/csv', 'text/plain'];
    this.files = [];
  }

  firstUpdated() {
    this.fetchFormats();
  }

  async fetchFormats() {
    const resp = await this.get('/schemes');

    if (resp.result) {
      this.schemes = resp.data;
    }
  }

  async startUpload() {
    const elements = Array.from(this.shadowRoot.querySelectorAll('tp-upload-item'));
    for (const element of elements) {
      await element.upload();
    }
  }

  filesSelected(e) {
    this.files = [ ...this.files, ...this.$.fakeUpload.files ];
  }

  removeFile(e) {
    const file = e.detail;
    this.files = this.files.filter(f => f !== file);
  }

  browse() {
    if (this.uploading) return;
    this.$.fakeUpload.click();
  }

  reset() {
    this.files = [];
  }
}

window.customElements.define('the-files', TheFiles);