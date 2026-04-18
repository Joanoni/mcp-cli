const fs = require('fs');
const path = require('path');

const src = path.resolve(__dirname, 'mcp_settings.json');
const dest = path.resolve(__dirname, '../../.roo/mcp.json');

fs.copyFileSync(src, dest);

const srcContents = fs.readFileSync(src, 'utf8');
const destContents = fs.readFileSync(dest, 'utf8');

if (srcContents === destContents) {
  console.log('✅ Success: .roo/mcp.json is up to date');
  process.exit(0);
} else {
  console.error('❌ Error: files differ after copy');
  process.exit(1);
}
