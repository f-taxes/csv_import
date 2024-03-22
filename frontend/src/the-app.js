/**
@license
Copyright (c) 2022 trading_peter
This program is available under Apache License Version 2.0
*/

import '@tp/tp-router/tp-router.js';
import './the-files.js';
import { LitElement, html, css } from 'lit';
import theme from './styles/theme.js';
import shared from './styles/shared.js';

class TheApp extends LitElement {
  static get styles() {
    return [
      theme,
      shared,
      css`
        :host {
          display: block;
          position: absolute;
          inset: 0;
        }

        .main {
          position: absolute;
          inset: 0;
          display: flex;
          flex-direction: column;
        }

        the-404 {
          flex: 1;
        }

        footer {
          display: flex;
          justify-content: space-between;
          align-items: center;
          background: var(--bg1);
          color: var(--hl1);
          padding: 15px 20px;
          margin-top: 3px;
        }

        footer > div {
          display: flex;
          align-items: center;
        }

        .heart {
          --tp-icon-width: 16px;
          --tp-icon-height: 16px;
        }

        .heart,
        a {
          padding: 0 5px;
        }
      `
    ];
  }

  render() {
    const { routeParams } = this;
    const p = routeParams || [];
    const page = p[0];

    return html`
      <tp-router @data-changed=${this.routeDataChanged}>
        <tp-route path="/config" data="files"></tp-route>
      </tp-router>
      
      <div class="main">
        ${page === 'files' ? html`<the-files .active=${page === 'files'}></the-files>` : null }
      </div>
    `;
  }

  static get properties() {
    return {
      // Data of the currently active route. Set by the router.
      route: { type: String, },

      // Params of the currently active route. Set by the router.
      routeParams: { type: Object },
    };
  }

  routeDataChanged(e) {
    this.route = e.detail;
    this.routeParams = this.route.split('-');
  }
}

window.customElements.define('the-app', TheApp);