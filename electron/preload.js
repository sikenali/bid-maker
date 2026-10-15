'use strict';

/** 预加载脚本:只暴露最小 API,渲染进程无 Node 权限。 */

const { contextBridge, ipcRenderer } = require('electron');

contextBridge.exposeInMainWorld('electronAPI', {
  isDesktop: true,
  getAppInfo: () => ipcRenderer.invoke('get-app-info'),
  getRuntimeInfo: () => ipcRenderer.invoke('get-runtime-info'),
  showAbout: () => ipcRenderer.invoke('show-about-dialog'),
});
