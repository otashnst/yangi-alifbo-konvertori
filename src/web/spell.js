// Uzbek (2026 Latin) spell checker on top of hunspell-asm + frequency list.
(function (root) {
  "use strict";
  const APOS_RE = /[’'‘`ʻ]/g;
  const WORD_RE = /[\p{L}ʼ’'‘`ʻ]+(?:-[\p{L}ʼ’'‘`ʻ]+)*/gu;
  const LETTER = /\p{L}/u;

  // suffixes written apart after numbers/times: "18:00 gaça", "5-sinf"
  const STANDALONE = new Set(["gaça", "dagi", "ning", "inçi", "nçi", "gaçа", "lardan", "larga", "dan", "ga", "da"]);
  function normApos(w) { return w.replace(APOS_RE, "ʼ"); }
  function isUpper(s) { return s === s.toUpperCase() && s !== s.toLowerCase(); }
  function caseLike(src, w) {
    if (isUpper(src) && src.length > 1) return w.toUpperCase();
    if (src[0] && isUpper(src[0])) return w[0].toUpperCase() + w.slice(1);
    return w;
  }
  // loose key: what a word "sounds like" when typed without the new letters
  function loose(w) {
    return w.toLowerCase().replace(/ʼ/g, "").replace(/ö/g, "o").replace(/ğ/g, "g")
      .replace(/ş/g, "s").replace(/ç/g, "c").replace(/x/g, "h").replace(/sh/g, "s").replace(/ch/g, "c");
  }
  function lev(a, b) {
    const m = a.length, n = b.length; if (Math.abs(m - n) > 3) return 9;
    let prev = Array.from({ length: n + 1 }, (_, j) => j), cur = new Array(n + 1);
    for (let i = 1; i <= m; i++) {
      cur[0] = i;
      for (let j = 1; j <= n; j++) cur[j] = Math.min(prev[j] + 1, cur[j - 1] + 1, prev[j - 1] + (a[i - 1] === b[j - 1] ? 0 : 1));
      [prev, cur] = [cur, prev];
    }
    return prev[n];
  }

  function Speller(hs, freqText) {
    this.hs = hs;
    this.rank = new Map();
    const lines = freqText.split("\n");
    for (let i = 0; i < lines.length; i++) if (lines[i] && !this.rank.has(lines[i])) this.rank.set(lines[i], i + 1);
    this.cache = new Map();
  }
  Speller.prototype.r = function (w) { return this.rank.get(w) || 1e9; };
  Speller.prototype.spell = function (w) { try { return this.hs.spell(w); } catch (e) { return true; } };
  Speller.prototype.known = function (w) {
    const l = w.toLowerCase();
    return this.spell(w) || this.spell(l) || this.rank.has(l);
  };

  // restore letters people usually type wrong: o->ö, g->ğ, u->ö, x<->h, s->ş, c->ç
  Speller.prototype.variants = function (lw) {
    const pos = [];
    for (let i = 0; i < lw.length; i++) {
      const c = lw[i], nx = lw[i + 1] || "";
      if (c === "o") pos.push([i, ["ö"]]);
      else if (c === "g") pos.push([i, ["ğ"]]);
      else if (c === "u") pos.push([i, ["ö"]]);
      else if (c === "x") pos.push([i, ["h"]]);
      else if (c === "h" && lw[i - 1] !== "s" && lw[i - 1] !== "c") pos.push([i, ["x"]]);
      else if (c === "s" && nx !== "h") pos.push([i, ["ş"]]);
      else if (c === "c") pos.push([i, ["ç"]]);
    }
    if (pos.length > 9) pos.length = 9;
    const out = new Map();
    const total = 1 << pos.length;
    for (let mask = 1; mask < total; mask++) {
      let bits = 0, arr = lw.split("");
      for (let k = 0; k < pos.length; k++) if (mask & (1 << k)) { arr[pos[k][0]] = pos[k][1][0]; bits++; }
      if (bits > 3) continue;
      let v = arr.join("");
      // a ş/ç typed as "sh"/"ch" leaves an h behind: shxar -> şahar handled by x->h; drop stray h after ş/ç
      v = v.replace(/şh/g, "ş").replace(/çh/g, "ç");
      if (v !== lw && !out.has(v) && this.spell(v)) out.set(v, bits);
    }
    return [...out.entries()].sort((a, b) => (this.r(a[0]) - this.r(b[0])) || (a[1] - b[1])).map(e => e[0]);
  };

  // returns null when the word looks fine, else {sugs, confident}
  Speller.prototype.check = function (word) {
    if (this.cache.has(word)) return this.cache.get(word);
    const res = this._check(word); this.cache.set(word, res); return res;
  };
  Speller.prototype._check = function (orig) {
    const w = normApos(orig).replace(/^[ʼ-]+|[ʼ-]+$/g, "");
    if (w.length < 3 || /\d/.test(w) || /[Ѐ-ӿ]/.test(w)) return null;
    if (isUpper(w)) return null;                       // abbreviations: KPI, MÇJ
    if (/\p{Lu}/u.test(w.slice(1).replace(/-\p{Lu}/gu, ""))) return null; // mixed case: kVli, MVt
    if (STANDALONE.has(w.toLowerCase())) return null;
    if (/[wqx]{2}|[^\p{L}ʼ-]/u.test(w)) return null;
    if (w.includes("-")) {
      if (this.spell(w)) return null;
      const parts = w.split("-").filter(Boolean);
      if (parts.every(p => p.length < 3 || this.known(p) || isUpper(p))) return null;
    }
    const lw = w.toLowerCase();
    const cap = w[0] !== lw[0];
    const ok = this.spell(w) || this.spell(lw);
    if (ok) return null;
    // obyekt / obʼekt, podyezd / podʼezd: both spellings in use
    const alt = lw.replace(/([bdtsvn])ye/g, "$1ʼe");
    if (alt !== lw && this.spell(alt)) return null;
    const rank = this.r(lw);
    const vars = this.variants(lw);
    if (vars.length) {
      const best = vars[0];
      // known-but-misspelled forms (xamma, boladi) lose to a far more common correct form
      // the dictionary rejects the word and a correctly spelled twin is more common (xamma, buyiça, boladi)
      if (rank === 1e9 || this.r(best) < rank) {
        return { sugs: vars.slice(0, 5).map(v => caseLike(w, v)), confident: !cap };
      }
    }
    if (rank !== 1e9) return null;                      // in frequency list: accept
    if (cap) return null;                                // probably a name
    let sugs = [];
    try { sugs = this.hs.suggest(lw); } catch (e) { sugs = []; }
    const lk = loose(lw);
    sugs = sugs.filter(s => !/[\s]/.test(s) && lev(loose(s), lk) <= Math.max(1, Math.floor(lw.length / 5)) + (loose(s) === lk ? 0 : 0));
    sugs.sort((a, b) => (loose(a) === lk ? 0 : 1) - (loose(b) === lk ? 0 : 1) || this.r(a) - this.r(b));
    const confident = sugs.length > 0 && loose(sugs[0]) === lk;
    return { sugs: sugs.slice(0, 5).map(v => caseLike(w, v)), confident };
  };

  // words: Map(word -> count). Returns sorted findings.
  Speller.prototype.analyze = function (words) {
    const out = [];
    for (const [word, count] of words) {
      const r = this.check(word);
      if (r) out.push({ word, count, sugs: r.sugs, confident: r.confident });
    }
    out.sort((a, b) => (b.confident - a.confident) || (b.sugs.length > 0) - (a.sugs.length > 0) || b.count - a.count || a.word.localeCompare(b.word));
    return out;
  };

  function collectWords(text, map) {
    let m; WORD_RE.lastIndex = 0;
    while ((m = WORD_RE.exec(text))) map.set(m[0], (map.get(m[0]) || 0) + 1);
    return map;
  }
  function escRe(s) { return s.replace(/[.*+?^${}()|[\]\\]/g, "\\$&"); }
  function applyFixes(text, fixes) {
    if (!fixes || !fixes.size) return text;
    if (!fixes._re) fixes._re = new RegExp("(?<![\\p{L}ʼ’'‘`ʻ-])(" + [...fixes.keys()].sort((a, b) => b.length - a.length).map(escRe).join("|") + ")(?![\\p{L}ʼ’'‘`ʻ])", "gu");
    return text.replace(fixes._re, m => fixes.get(m) ?? m);
  }

  root.UzSpell = { Speller, collectWords, applyFixes, LETTER };
})(typeof window !== "undefined" ? window : globalThis);
