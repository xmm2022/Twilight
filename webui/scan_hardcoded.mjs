import fs from "fs";
import path from "path";

function walk(dir, out = []) {
  for (const e of fs.readdirSync(dir, { withFileTypes: true })) {
    const p = path.join(dir, e.name);
    if (e.isDirectory()) walk(p, out);
    else if (/\.(tsx?|jsx?)$/.test(e.name)) out.push(p);
  }
  return out;
}

// strip // line comments and /* */ block comments (rough; ignores strings containing // )
function stripComments(src) {
  let out = "";
  let i = 0;
  const n = src.length;
  let mode = "code"; // code | line | block | sq | dq | tpl
  while (i < n) {
    const c = src[i], c2 = src[i + 1];
    if (mode === "code") {
      if (c === "/" && c2 === "/") { mode = "line"; i += 2; continue; }
      if (c === "/" && c2 === "*") { mode = "block"; i += 2; continue; }
      if (c === "'") { mode = "sq"; out += c; i++; continue; }
      if (c === '"') { mode = "dq"; out += c; i++; continue; }
      if (c === "`") { mode = "tpl"; out += c; i++; continue; }
      out += c; i++; continue;
    }
    if (mode === "line") { if (c === "\n") { mode = "code"; out += c; } i++; continue; }
    if (mode === "block") { if (c === "*" && c2 === "/") { mode = "code"; i += 2; } else i++; continue; }
    if (mode === "sq") { out += c; if (c === "\\") { out += c2; i += 2; continue; } if (c === "'") mode = "code"; i++; continue; }
    if (mode === "dq") { out += c; if (c === "\\") { out += c2; i += 2; continue; } if (c === '"') mode = "code"; i++; continue; }
    if (mode === "tpl") { out += c; if (c === "\\") { out += c2; i += 2; continue; } if (c === "`") mode = "code"; i++; continue; }
  }
  return out;
}

const cjk = /[一-鿿]/;
const files = walk("src");
const rows = [];
for (const f of files) {
  const src = fs.readFileSync(f, "utf8");
  const stripped = stripComments(src);
  const lines = stripped.split("\n");
  let count = 0;
  for (const l of lines) if (cjk.test(l)) count++;
  if (count > 0) rows.push([f.replace(/\\/g, "/"), count]);
}
rows.sort((a, b) => b[1] - a[1]);
console.log("=== files with non-comment CJK (likely hardcoded strings), by count ===");
let total = 0;
for (const [f, c] of rows) { console.log(String(c).padStart(4), f); total += c; }
console.log("\nTOTAL non-comment CJK lines:", total, "across", rows.length, "files");
