"""Assemble app/index.html from page.html + inline scripts (run from src/)."""
import pathlib
root = pathlib.Path(__file__).resolve().parent.parent
web = root / "web"
h = (web / "page.html").read_text(encoding="utf-8")
parts = {
    "<!--JSZIP-->": "<script>" + (web / "node_modules/jszip/dist/jszip.min.js").read_text(encoding="utf-8") + "</script>",
    "<!--HUNSPELL-->": "<script>" + (web / "hunspell.bundle.js").read_text(encoding="utf-8") + "</script>",
    "<!--SPELL-->": "<script>" + (web / "spell.js").read_text(encoding="utf-8") + "</script>",
    "<!--APP-->": (web / "app.js").read_text(encoding="utf-8"),
}
for k, v in parts.items():
    assert k in h, k
    h = h.replace(k, v, 1)
(root / "app" / "index.html").write_text(h, encoding="utf-8")
print("app/index.html", len(h))
