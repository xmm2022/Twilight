import fs from "fs";
const dir = "src/locales";
const base = JSON.parse(fs.readFileSync(`${dir}/basic.json`, "utf8"));
const files = ["zh-Hans.json", "zh-Hant.json", "en-US.json"];
const loaded = {};
for (const f of files) loaded[f] = JSON.parse(fs.readFileSync(`${dir}/${f}`, "utf8"));

function leaves(obj, prefix = "", out = {}) {
  for (const k of Object.keys(obj)) {
    const v = obj[k], key = prefix ? `${prefix}.${k}` : k;
    if (v && typeof v === "object") leaves(v, key, out);
    else out[key] = v;
  }
  return out;
}
const ph = (s) => [...String(s).matchAll(/\{([a-zA-Z0-9_]+)\}/g)].map((m) => m[1]).sort();
const Lb = leaves(base);
for (const f of files) {
  const Lf = leaves(loaded[f]);
  console.log(`\n=== placeholder mismatches: basic vs ${f} ===`);
  let n = 0;
  for (const k of Object.keys(Lf)) {
    const a = JSON.stringify(ph(Lb[k] ?? "")), b = JSON.stringify(ph(Lf[k]));
    if (a !== b) { console.log(`  ${k}\n    basic: ${a}\n    ${f}:  ${b}`); n++; }
  }
  if (!n) console.log("  (none)");
}
