// Replacement for nanoid@2 used by hunspell-asm only to name an in-memory folder.
// Uses the browser's CSPRNG; validates size, so the nanoid@<3.3.18 advisories do not apply.
module.exports = function nanoid(size) {
  var n = (typeof size === "number" && isFinite(size)) ? Math.max(1, Math.min(128, Math.floor(size))) : 21;
  var abc = "ModuleSymbhasOwnPr0123456789ABCDEFGHNRVfgctiUvz_KqYTJkLxpZXIjQW";
  var b = new Uint8Array(n); crypto.getRandomValues(b);
  var s = ""; for (var i = 0; i < n; i++) s += abc[b[i] & 63];
  return s;
};
