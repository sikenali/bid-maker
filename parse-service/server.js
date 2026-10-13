'use strict';

/** 解析微服务:Go 主后端经 HTTP 调用本服务.
 *
 *   POST /parse  multipart(file=招标文件) -> BidParsingOutput JSON
 *   GET  /health -> {"status": "ok"}
 *
 * 启动:  node server.js (PORT 环境变量,默认 8000)
 * Go 端配置: PYTHON_SERVICE=http://localhost:8000 (变量名保持兼容,值指向本服务)
 */

const express = require('express');
const multer = require('multer');
const { parseBidFile } = require('./pipeline');
const { renderDocx } = require('./exporter');

const PORT = parseInt(process.env.PORT || '8000', 10);
const ALLOWED_EXTS = new Set(['.docx', '.xlsx', '.xlsm', '.pptx', '.pdf']);
const MAX_BYTES = 50 * 1024 * 1024;

const app = express();
const upload = multer({ storage: multer.memoryStorage(), limits: { fileSize: MAX_BYTES } });

app.get('/health', (req, res) => {
  res.json({ status: 'ok', service: 'bid-parse-service', schema: 'v1', engine: 'node-officeparser' });
});

app.post('/parse', upload.single('file'), async (req, res) => {
  try {
    if (!req.file || !req.file.buffer || !req.file.buffer.length) {
      return res.status(400).json({ error: '空文件' });
    }
    const lower = (req.file.originalname || '').toLowerCase();
    const ext = Object.keys({ '.docx': 1, '.xlsx': 1, '.xlsm': 1, '.pptx': 1, '.pdf': 1 })
      .find((e) => lower.endsWith(e));
    if (!ext) {
      return res.status(400).json({ error: `不支持的文件格式:仅支持 .docx/.xlsx/.pptx/.pdf` });
    }
    const result = await parseBidFile(req.file.buffer, {
      titleHint: req.body.title || req.query.title || '',
      bidNoHint: req.body.bid_no || req.query.bid_no || '',
      filenameHint: req.file.originalname || '',
    });
    res.json(result);
  } catch (e) {
    res.status(500).json({ error: `解析失败: ${e.message || e}` });
  }
});

app.use(express.json({ limit: '10mb' }));

app.post('/export', (req, res) => {
  try {
    const { title, bid_no, export_time, sections } = req.body || {};
    if (!Array.isArray(sections) || !sections.length) {
      return res.status(400).json({ error: 'sections 非空数组必填' });
    }
    const buf = renderDocx({ title, bid_no, export_time, sections });
    res.set('Content-Type', 'application/vnd.openxmlformats-officedocument.wordprocessingml.document');
    res.send(buf);
  } catch (e) {
    res.status(500).json({ error: `导出失败: ${e.message || e}` });
  }
});

// multer 文件超限错误转为 413
// eslint-disable-next-line no-unused-vars
app.use((err, req, res, next) => {
  if (err && err.code === 'LIMIT_FILE_SIZE') {
    return res.status(413).json({ error: '文件超过 50MB 上限' });
  }
  res.status(500).json({ error: String((err && err.message) || err) });
});

app.listen(PORT, '0.0.0.0', () => {
  console.log(`bid-parse-service (node) listening on :${PORT}`);
});
