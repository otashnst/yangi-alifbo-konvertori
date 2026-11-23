
(function(){
const TOKEN=new URLSearchParams(location.search).get("t")||"";
const api=(p,o={})=>fetch(p,{...o,cache:"no-store",headers:{...(o.headers||{}),"X-Token":TOKEN}});
const APOS="‘'ʻ`’ʼ";
const MAP={O:"Ö",o:"ö",G:"Ğ",g:"ğ"};
const reOG=new RegExp("([OoGg])["+APOS+"]","g");
const reProtect=/(https?:\/\/\S+|www\.\S+|[\w.+-]+@[\w-]+\.[\w.-]+)/g;


// ---- Kirill -> yangi lotin ----
const C1={"а":"a","б":"b","в":"v","г":"g","д":"d","ж":"j","з":"z","и":"i","й":"y","к":"k","л":"l","м":"m","н":"n","о":"o","п":"p","р":"r","с":"s","т":"t","у":"u","ф":"f","х":"x","ч":"ç","ш":"ş","щ":"ş","ы":"i","э":"e","ў":"ö","қ":"q","ғ":"ğ","ҳ":"h","ё":"yo","ю":"yu","я":"ya"};
const CV="аеёиоуэюяўыАЕЁИОУЭЮЯЎЫ";
const isCyr=c=>/[Ѐ-ӿ]/.test(c||"");
const isUp=c=>!!c&&c!==c.toLowerCase();
const isLetter=c=>!!c&&/\p{L}/u.test(c);
function cyr2lat(s,counts,prev){
  if(!/[Ѐ-ӿ]/.test(s))return s;
  let out="";
  for(let i=0;i<s.length;i++){
    const ch=s[i],lo=ch.toLowerCase();
    if(!isCyr(ch)){out+=ch;continue}
    const p=i?s[i-1]:(prev||""),n=s[i+1]||"";
    let r;
    if(lo==="е"){r=(!isLetter(p)||CV.includes(p)||"ъьЪЬ".includes(p))?"ye":"e"}
    else if(lo==="ц"){r=(isLetter(p)&&CV.includes(p))?"ts":"s"}
    else if(lo==="ъ"){r=("еёюяЕЁЮЯ".includes(n)||"ўЎ".includes(p))?"":"’"}
    else if(lo==="ь"){r=""}
    else r=C1[lo];
    if(r===undefined){out+=ch;continue}
    counts.K++;
    if(isUp(ch)&&r){
      const allCaps=(isUp(n)&&isLetter(n))||(isUp(p)&&isLetter(p)&&!(isLetter(n)&&!isUp(n)));
      r=r.length>1&&!allCaps?r[0].toUpperCase()+r.slice(1):r.toUpperCase();
    }
    out+=r;
  }
  return out;
}

function skipList(){
  return document.getElementById("skip").value.split(",").map(s=>s.trim()).filter(Boolean);
}
function escRe(s){return s.replace(/[.*+?^${}()|[\]\\]/g,"\\$&")}

// convert a plain string; counts accumulates per-letter totals
function convertPlain(s,counts){
  return s.replace(reOG,(m,c)=>{counts[MAP[c].toUpperCase()]++;return MAP[c]})
    .replace(/S[Hh]/g,()=>{counts["Ş"]++;return "Ş"})
    .replace(/s[hH]/g,()=>{counts["Ş"]++;return "ş"})
    .replace(/C[Hh]/g,()=>{counts["Ç"]++;return "Ç"})
    .replace(/c[hH]/g,()=>{counts["Ç"]++;return "ç"});
}
function convert(s,counts,skip,prev){
  // protect URLs/emails and user-listed words
  const parts=[];let pat=reProtect.source;
  if(skip.length) pat+="|(?<![\\p{L}])(?:"+skip.map(escRe).join("|")+")(?![\\p{L}])";
  const re=new RegExp(pat,"gu");let last=0,m;
  while((m=re.exec(s))){parts.push([s.slice(last,m.index),true],[m[0],false]);last=m.index+m[0].length;if(!m[0].length)re.lastIndex++}
  parts.push([s.slice(last),true]);
  const doCyr=document.getElementById("cyr").checked;let pc=prev||"",res="";
  for(const [t,c] of parts){let o=t;if(c){o=convertPlain(t,counts);if(doCyr)o=cyr2lat(o,counts,pc)}res+=o;if(t)pc=t[t.length-1]}
  return res;
}
function newCounts(){return {"Ö":0,"Ğ":0,"Ş":0,"Ç":0,K:0}}

// XML helpers
function dec(s){return s.replace(/&(#x[0-9a-f]+|#\d+|amp|lt|gt|quot|apos);/gi,(m,e)=>{
  if(e[0]==="#")return String.fromCodePoint(e[1]==="x"||e[1]==="X"?parseInt(e.slice(2),16):parseInt(e.slice(1),10));
  return {amp:"&",lt:"<",gt:">",quot:'"',apos:"'"}[e.toLowerCase()]})}
function enc(s){return s.replace(/&/g,"&amp;").replace(/</g,"&lt;").replace(/>/g,"&gt;")}

// convert text nodes in an OOXML part; joins letters split across runs inside one paragraph
function convertXml(xml,tag,counts,skip,words,ctx2){
  const re=new RegExp("(<"+tag+"(?:\\s[^>]*)?>)([^<]*)(</"+tag+">)","g");
  const nodes=[];let m;
  while((m=re.exec(xml)))nodes.push({start:m.index,end:m.index+m[0].length,open:m[1],text:dec(m[2]),close:m[3]});
  if(!nodes.length)return xml;
  const breakRe=/<\/w:p>|<\/a:p>|<\/si>|<\/is>|<w:br\b|<w:tab\b|<a:br\b|<w:cr\b/;
  for(let i=0;i<nodes.length-1;i++){
    const a=nodes[i],b=nodes[i+1];
    if(!a.text||!b.text)continue;
    if(breakRe.test(xml.slice(a.end,b.start)))continue;
    const l=a.text.slice(-1),f=b.text[0];
    if("OoGg".includes(l)&&APOS.includes(f)){a.text=a.text.slice(0,-1)+l+f;b.text=b.text.slice(1)}
    else if("SsCc".includes(l)&&"hH".includes(f)){a.text=a.text+f;b.text=b.text.slice(1)}
  }
  let prevNode=null;const groups=[];let g=null;
  for(const n of nodes){
    const joined=prevNode&&!breakRe.test(xml.slice(prevNode.end,n.start));
    const ctx=joined&&prevNode.text?prevNode.text.slice(-1):"";
    prevNode=n;
    n.out=convert(n.text,counts,skip,ctx);
    if(n.text!==n.out)collectWords(n.text,n.out,words);
    if(!joined||!g){g=[];groups.push(g)}
    g.push(n);
  }
  for(const grp of groups){
    const full=grp.map(n=>n.out).join("");
    if(ctx2.all)UzSpell.collectWords(full,ctx2.all);
    if(ctx2.fixes&&ctx2.fixes.size)fixGroup(grp,full,ctx2.fixes);
  }
  let out="",pos=0;
  for(const n of nodes){out+=xml.slice(pos,n.start)+n.open+enc(n.out)+n.close;pos=n.end}
  return out+xml.slice(pos);
}
// apply word fixes to a paragraph whose text is split over several runs
function fixGroup(grp,full,fixes){
  UzSpell.applyFixes("",fixes);
  const ms=[...full.matchAll(fixes._re)];if(!ms.length)return;
  const starts=[];let acc=0;for(const n of grp){starts.push(acc);acc+=n.out.length}
  for(let k=ms.length-1;k>=0;k--){
    const m=ms[k],a=m.index,b=a+m[0].length,rep=fixes.get(m[0]);if(rep==null)continue;
    let first=true;
    for(let i=grp.length-1;i>=0;i--){
      const s0=starts[i],s1=s0+grp[i].out.length;
      if(s1<=a||s0>=b)continue;
      const lo=Math.max(a,s0)-s0,hi=Math.min(b,s1)-s0;
      const isFirst=s0<=a;
      grp[i].out=grp[i].out.slice(0,lo)+(isFirst?rep:"")+grp[i].out.slice(hi);
    }
  }
}
function collectWords(before,after,words){
  if(words.size>=60)return;
  const tok=after.split(/[^\p{L}\p{M}’'‘ʻ-]+/u);
  for(const w of tok){if(w!==""&&(/[ÖöĞğŞşÇç]/.test(w)||!before.includes(w))&&!/[\u0400-\u04FF]/.test(w)&&words.size<60)words.add(w)}
}

const PARTS={
  docx:[[/^word\/(document|footnotes|endnotes|comments|header\d*|footer\d*)\.xml$/,"w:t"],[/^word\/(charts|diagrams)\/.*\.xml$/,"a:t"],[/^word\/document\.xml$/,"a:t"]],
  xlsx:[[/^xl\/sharedStrings\.xml$/,"t"],[/^xl\/worksheets\/sheet\d+\.xml$/,"t"],[/^xl\/(charts|drawings)\/.*\.xml$/,"a:t"],[/^xl\/comments\d*\.xml$/,"t"]],
  pptx:[[/^ppt\/(slides|notesSlides|diagrams|charts)\/[^/]+\.xml$/,"a:t"]]
};

async function processFile(file,fixes){
  const ext=file.name.split(".").pop().toLowerCase();
  const counts=newCounts(),words=new Set(),skip=skipList();
  const ctx2={fixes,all:new Map()};
  const base=file.name.replace(/\.[^.]+$/,"");
  const outName=base+"_yangi_alifbo."+ext;
  if(ext==="txt"){
    const t=await file.text();let o=convert(t,counts,skip);collectWords(t,o,words);
    UzSpell.collectWords(o,ctx2.all);if(fixes&&fixes.size)o=UzSpell.applyFixes(o,fixes);
    return {name:outName,blob:new Blob([o],{type:"text/plain"}),counts,words,all:ctx2.all};
  }
  if(!PARTS[ext])throw new Error("Bu fayl turi qo‘llab-quvvatlanmaydi. .docx, .xlsx, .pptx yoki .txt fayl tanlang.");
  if(typeof JSZip==="undefined")throw new Error("Kutubxona yuklanmadi. Sahifani yangilab ko‘ring.");
  let zip;
  try{zip=await JSZip.loadAsync(file)}catch(e){throw new Error("Faylni ochib bo‘lmadi. U buzilgan yoki eski .doc/.xls formatida bo‘lishi mumkin — avval Word/Excel’da .docx/.xlsx qilib saqlang.")}
  for(const path of Object.keys(zip.files)){
    if(zip.files[path].dir)continue;
    const rules=PARTS[ext].filter(([re])=>re.test(path));
    if(!rules.length)continue;
    let xml=await zip.file(path).async("string");const orig=xml;
    for(const [,tag] of rules)xml=convertXml(xml,tag,counts,skip,words,ctx2);
    if(xml!==orig)zip.file(path,xml);
  }
  const blob=await zip.generateAsync({type:"blob",compression:"DEFLATE",compressionOptions:{level:6}});
  return {name:outName,blob,counts,words,all:ctx2.all};
}

// downloads capability
let dl=null;
window.__saves={};window.__saveDone=(id,status,path)=>{const f=window.__saves[id];delete window.__saves[id];if(f)f({status,path})};
function blobToB64(b){return new Promise((res,rej)=>{const fr=new FileReader();fr.onload=()=>res(String(fr.result).split(",")[1]||"");fr.onerror=rej;fr.readAsDataURL(b)})}
window.addEventListener("dragover",e=>e.preventDefault());window.addEventListener("drop",e=>e.preventDefault());
if(!window.saveFile){setInterval(()=>api("/ping").catch(()=>{}),2000);api("/ping").catch(()=>{})}
if(window.claude&&window.claude.use){window.claude.use("downloads").then(d=>{dl=d}).catch(()=>{})}

async function saveBlob(name,blob,btn,msg){
  if(typeof window.saveFile==="function"){
    btn.disabled=true;msg.className="note";msg.textContent="";
    try{
      const b64=await blobToB64(blob);
      const res=await new Promise(res=>{const id=String(Date.now())+Math.random();window.__saves[id]=res;window.saveFile(id,name,b64)});
      if(res.status==="saved"){msg.textContent="Saqlandi: "+res.path+" ";const o=el("button","mini","Papkani ochish");o.onclick=()=>window.openFolder(res.path);msg.append(o)}
      else if(res.status==="cancel"){msg.textContent="Bekor qilindi."}
      else{msg.className="err";msg.textContent="Saqlab bo‘lmadi: "+res.path}
    }catch(e){msg.className="err";msg.textContent="Saqlab bo‘lmadi."}
    btn.disabled=false;return}
  if(!dl){const a=document.createElement("a");a.href=URL.createObjectURL(blob);a.download=name;document.body.append(a);a.click();a.remove();setTimeout(()=>URL.revokeObjectURL(a.href),5000);msg.textContent="Saqlandi: "+name+" (Yuklanmalar papkasida)";return}
  btn.disabled=true;msg.textContent="";msg.className="note";
  try{await dl.save({filename:name,data:blob});msg.textContent="Saqlandi: "+name}
  catch(e){msg.className=e&&e.code==="declined"?"note":"err";
    msg.textContent=e&&e.code==="declined"?"Bekor qilindi.":e&&e.code==="rate_limited"?"Biroz kutib, qayta bosing.":"Saqlab bo‘lmadi ("+(e&&e.code||"xato")+")."}
  btn.disabled=false;
}

function el(tag,cls,text){const e=document.createElement(tag);if(cls)e.className=cls;if(text!=null)e.textContent=text;return e}
const $=id=>document.getElementById(id);

// ---- imlo lug'ati ----
let speller=null,spellState="loading",loadedDictVer=null;
const spellOn=()=>$("spellOn").checked;
function spellBadge(){
  const b=$("spellState");
  b.textContent=spellState==="ready"?"lug‘at tayyor":spellState==="loading"?"lug‘at yuklanmoqda…":"mavjud emas";
  b.className="badge "+spellState;
  if(spellState==="off"){$("spellOn").checked=false;$("spellOn").disabled=true}
}
async function getDictInfo(){try{const r=await api("dict/info");if(!r.ok)return null;return await r.json()}catch(e){return null}}
async function initSpell(){
  try{
    if(!window.HunspellAsm||!window.UzSpell)throw new Error("no engine");
    const info=await getDictInfo();
    const get=(f,bin)=>api("dict/"+f).then(r=>{if(!r.ok)throw new Error(f);return bin?r.arrayBuffer():r.text()});
    const [aff,dic,freq]=await Promise.all([get("uz_UZ-2026.aff",1),get("uz_UZ-2026.dic",1),get("freq.txt",0)]);
    const f=await HunspellAsm.loadModule();
    const h=f.create(f.mountBuffer(new Uint8Array(aff),"uz"+Date.now()+".aff"),f.mountBuffer(new Uint8Array(dic),"uz"+Date.now()+".dic"));
    speller=new UzSpell.Speller(h,freq);spellState="ready";loadedDictVer=info?info.version:null;
  }catch(e){if(!speller)spellState="off"}
  spellBadge();liveSpell();
}
let spellReady=initSpell();
function findErrors(all){
  if(!speller)return [];
  const skip=new Set(skipList().map(x=>x.toLowerCase()));
  const m=new Map();for(const [w,c] of all)if(!skip.has(w.toLowerCase()))m.set(w,c);
  return speller.analyze(m);
}

// ---- lug'at yangilanishi (faqat dastur ichida) ----
const DICT_MSG={disabled:"Internet o‘chirilgan — mavjud lug‘at ishlatilmoqda",idle:"Yangilanish tekshirilmoqda…",checking:"Yangilanish tekshirilmoqda…",latest:"Eng so‘nggi versiya",updated:"Yangilandi",offline:"Internet yo‘q — joriy lug‘at ishlatilmoqda",error:"Yangilab bo‘lmadi",app_old:"Dasturni yangilash kerak"};
function showDict(i){
  if(!i){$("dictVer").textContent="ichki";$("dictMsg").textContent="Dastur ichidagi lug‘at";$("dictBtn").hidden=true;return}
  $("dictVer").textContent=i.version;
  let m=DICT_MSG[i.status]||"";if(i.message&&/^(updated|error|app_old)$/.test(i.status))m+=": "+i.message;
  if(i.status==="latest"&&i.source==="updated")m="Eng so‘nggi versiya (yangilangan)";
  $("dictMsg").textContent=m;$("dictBtn").disabled=i.status==="checking";$("dictBtn").hidden=i.status==="disabled";
}
async function afterDictChange(i){
  if(i&&i.status==="updated"&&speller&&loadedDictVer&&i.version!==loadedDictVer){
    spellState="loading";spellBadge();spellReady=initSpell();await spellReady;
    if(S.r&&spellOn()){S.found=findErrors(S.r.all);renderTable()}
  }
}
(async()=>{
  let i=await getDictInfo();showDict(i);if(!i)return;
  for(let k=0;k<30&&(i.status==="idle"||i.status==="checking");k++){await new Promise(r=>setTimeout(r,2000));i=await getDictInfo()||i;showDict(i)}
  afterDictChange(i);
})();
$("dictBtn").onclick=async()=>{
  $("dictBtn").disabled=true;$("dictMsg").textContent=DICT_MSG.checking;
  let i=null;try{const r=await api("dict/update",{method:"POST"});i=await r.json()}catch(e){}
  if(!i){$("dictMsg").textContent="Yangilab bo‘lmadi";$("dictBtn").disabled=false;return}
  showDict(i);afterDictChange(i);
};

// ---- dasturni yangilash ----
async function getAppInfo(){try{const r=await api("app/update-info");if(!r.ok)return null;return await r.json()}catch(e){return null}}
function showApp(i){
  const box=$("updBox");if(!i||!/^(available|downloading|installing)$/.test(i.status)){box.hidden=true;return}
  box.hidden=false;$("updTitle").textContent=i.version+" versiya chiqdi";$("updNotes").textContent=i.notes||"";
  const b=$("updBtn"),m=$("updMsg");
  if(i.status==="available"){b.disabled=false;b.textContent="Yangilash";m.textContent=i.message||""}
  else if(i.status==="downloading"){b.disabled=true;b.textContent="Yuklanmoqda…";m.textContent=""}
  else{b.disabled=true;b.textContent="O‘rnatilmoqda…";m.textContent="Dastur yopiladi va yangi versiyada qayta ochiladi."}
}
(async()=>{
  let i=await getAppInfo();if(!i)return;
  for(let k=0;k<30&&(i.status==="idle"||i.status==="checking");k++){await new Promise(r=>setTimeout(r,2000));i=await getAppInfo()||i}
  showApp(i);
})();
$("updBtn").onclick=async()=>{
  showApp({...(await getAppInfo()||{}),status:"downloading"});
  let i=null;try{const r=await api("app/update",{method:"POST"});i=await r.json()}catch(e){}
  showApp(i||{status:"available",message:"Yangilab bo‘lmadi"});
};

// ---- sozlamalar: internetsiz rejim ----
async function loadNet(){
  let st=null;try{const r=await api("app/settings");if(r.ok)st=await r.json()}catch(e){}
  const box=$("netBox");if(!st){box.hidden=true;return null}
  box.hidden=false;$("netOff").checked=st.offline;$("netOff").disabled=st.policy;$("netNote").hidden=!st.policy;
  return st;
}
loadNet();
$("netOff").onchange=async()=>{
  const want=$("netOff").checked;
  try{await api("app/settings",{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify({offline:want})})}catch(e){}
  await loadNet();showDict(await getDictInfo());
  if(want)$("updBox").hidden=true;
};

// ---- qadamlar ----
const VIEWS={1:"v1",2:"v2",3:"v3",4:"v4",text:"vtext"};
const S={file:null,r:null,found:null,rows:[],fixes:new Map(),maxStep:1,saved:false,blobKey:null,blob:null};
function show(step){
  for(const [k,id] of Object.entries(VIEWS))$(id).hidden=String(k)!==String(step);
  document.querySelectorAll(".steps button").forEach(b=>{
    const s=b.dataset.step,cur=String(s)===String(step);
    if(cur)b.setAttribute("aria-current","step");else b.removeAttribute("aria-current");
    if(s==="text")return;
    const n=+s,done=n<S.maxStep||(n===4&&S.saved);
    b.disabled=n>S.maxStep;b.classList.toggle("done",done&&!cur);
    b.querySelector(".dot").textContent=done&&!cur?"✓":String(n);
  });
  document.querySelector(".main").scrollTop=0;
}
document.querySelectorAll(".steps button").forEach(b=>b.onclick=()=>{if(!b.disabled)show(b.dataset.step)});
document.addEventListener("click",e=>{const g=e.target.closest("[data-go]");if(!g)return;const v=g.dataset.go;if(v==="new")reset();else show(+v)});
function reset(){
  Object.assign(S,{file:null,r:null,found:null,rows:[],fixes:new Map(),maxStep:1,saved:false,blobKey:null,blob:null});
  $("saveMsg").textContent="";show(1);
}

const KIND={docx:"Word hujjati",xlsx:"Excel jadvali",pptx:"PowerPoint taqdimoti",txt:"Matn fayli"};
function fmtSize(n){return n>1048576?(n/1048576).toFixed(1)+" MB":Math.max(1,Math.round(n/1024))+" KB"}
function fillCards(){
  const r=S.r,f=S.file,ext=f.name.split(".").pop().toLowerCase();
  const total=Object.entries(r.counts).reduce((a,[k,v])=>a+v,0);
  for(const c of document.querySelectorAll("[data-fcard]")){
    c.textContent="";
    c.insertAdjacentHTML("beforeend",'<svg width="34" height="40" viewBox="0 0 34 40" aria-hidden="true"><path d="M3 3h19l9 9v25H3z" fill="#e9eef6" stroke="#1d3a6e" stroke-width="2"></path><path d="M22 3v9h9" fill="none" stroke="#1d3a6e" stroke-width="2"></path><path d="M9 20h16M9 26h16M9 32h10" stroke="#1d3a6e" stroke-width="2"></path></svg>');
    const meta=el("div","meta");meta.append(el("span","name",f.name),el("span","sub",(total?total+" ta harf o‘girildi":"Almashtiriladigan harf topilmadi")+" · "+(KIND[ext]||"Fayl")+" · "+fmtSize(f.size)));
    const ch=el("div","chips");for(const [k,v] of Object.entries(r.counts)){if(k==="K"){if(v)ch.append(el("span","","Kirill "+v))}else ch.append(el("span","",k+" "+v))}
    c.append(meta,ch);
  }
  const w=$("v2words");w.textContent="";
  if(r.words.size){const d=el("div","words");d.append(el("span","","O‘zgargan so‘zlardan namuna: "));[...r.words].slice(0,40).forEach((x,i)=>{if(i)d.append(", ");d.append(el("b","",x))});w.append(d)}
  else w.append(el("div","note","Faylda eski alifbodagi harflar topilmadi. U allaqachon yangi alifboda yozilgan bo‘lishi mumkin."));
  $("sumA").textContent=total;
}

function renderTable(){
  const t=$("v3tbl");t.textContent="";S.rows=[];
  const found=S.found||[];
  const sure=found.filter(f=>f.confident&&f.sugs.length),maybe=found.filter(f=>!f.confident&&f.sugs.length),unk=found.filter(f=>!f.sugs.length);
  const head=el("div","trow thead");head.append(el("span"),el("span","","Hujjatda"),el("span"),el("span","","Tuzatish"),el("span","n","Soni"));t.append(head);
  if(!sure.length&&!maybe.length){t.append(el("div","empty","✓ Imlo xatosi topilmadi"))}
  let first=true;
  function group(title,cls,list,checked){
    if(!list.length)return;
    const bar=el("div","tbar");bar.append(el("span","gtitle "+cls,title+" · "+list.length));
    if(first&&sure.length+maybe.length>1){const a=el("span","row");const all=el("button","mini","Hammasini belgilash"),none=el("button","mini","Belgilarni olib tashlash");
      all.onclick=()=>{S.rows.forEach(r=>{r.cb.checked=true;r.row.classList.add("on")});updNext()};none.onclick=()=>{S.rows.forEach(r=>{r.cb.checked=false;r.row.classList.remove("on")});updNext()};a.append(all,none);bar.append(a)}
    first=false;t.append(bar);
    for(const f of list){
      const row=el("label","trow"+(checked?" on":""));const cb=el("input");cb.type="checkbox";cb.checked=checked;
      let pick;
      if(f.sugs.length>1){pick=el("select");for(const s of f.sugs){const o=el("option","",s);o.value=s;pick.append(o)}pick.onclick=e=>e.preventDefault&&null}
      else pick=el("span","fix",f.sugs[0]);
      row.append(cb,el("span","w",f.word),el("span","ar","→"),pick,el("span","n",String(f.count)));
      cb.onchange=()=>{row.classList.toggle("on",cb.checked);updNext()};
      if(pick.tagName==="SELECT")pick.onchange=updNext;
      t.append(row);S.rows.push({f,cb,pick,row});
    }
  }
  group("Aniq xatolar","sure",sure,true);
  group("Tekshirib ko‘ring","maybe",maybe,false);
  if(unk.length){const ft=el("div","tfoot");ft.append(el("span","","Lug‘atda topilmadi: "));unk.forEach((f,i)=>{if(i)ft.append(", ");ft.append(el("b","",f.word))});ft.append(el("span",""," — ism, atama yoki xato bo‘lishi mumkin; o‘zgartirilmaydi."));t.append(ft)}
  updNext();
}
function getFixes(){const m=new Map();for(const r of S.rows)if(r.cb.checked)m.set(r.f.word,r.pick.tagName==="SELECT"?r.pick.value:r.f.sugs[0]);return m}
function updNext(){const n=getFixes().size;$("v3next").textContent=n?(n+" ta tuzatish bilan saqlash →"):"Tuzatishsiz saqlash →"}

async function handleFile(files){
  const msg=$("v1msg");msg.hidden=true;
  const f=files[0];if(!f)return;
  const drop=$("drop"),lbl=drop.querySelector("strong"),old=lbl.textContent;
  lbl.textContent="O‘girilmoqda…";drop.setAttribute("aria-busy","true");
  try{
    const r=await processFile(f,null);
    Object.assign(S,{file:f,r,found:null,rows:[],fixes:new Map(),saved:false,blobKey:null,blob:null});
    if(spellOn()){lbl.textContent="Imlo tekshirilmoqda…";await spellReady;if(speller)S.found=findErrors(r.all)}
    fillCards();renderTable();
    $("v2next").textContent=S.found?"Imlo tekshiruviga o‘tish →":"Saqlashga o‘tish →";
    S.maxStep=2;show(2);
    if(files.length>1){msg.hidden=false;msg.className="note";msg.textContent="Bir vaqtda bitta fayl o‘giriladi. Birinchisi olindi: "+f.name}
  }catch(e){msg.hidden=false;msg.className="err";msg.textContent=e.message}
  finally{lbl.textContent=old;drop.removeAttribute("aria-busy")}
}
$("v2next").onclick=()=>{if(S.found){S.maxStep=Math.max(S.maxStep,3);show(3)}else goSave(new Map())};
$("v3next").onclick=()=>goSave(getFixes());
function goSave(fixes){
  S.fixes=fixes;S.saved=false;S.maxStep=4;
  let n=0;for(const r of S.rows)if(fixes.has(r.f.word))n+=r.f.count;
  $("sumS").textContent=n;$("saveMsg").textContent="";
  show(4);doSave();
}
async function doSave(){
  const btn=$("saveBtn"),msg=$("saveMsg");
  const key=[...S.fixes].join("|");
  if(S.blobKey!==key){
    msg.className="note";msg.textContent="Tayyorlanmoqda…";btn.disabled=true;
    try{S.blob=S.fixes.size?(await processFile(S.file,S.fixes)).blob:S.r.blob;S.blobKey=key}
    catch(e){msg.className="err";msg.textContent=e.message;btn.disabled=false;return}
    btn.disabled=false;msg.textContent="";
  }
  await saveBlob(S.r.name,S.blob,btn,msg);
  if(/^Saqlandi/.test(msg.textContent)){S.saved=true;msg.classList.add("okmsg");show(4)}
}
$("saveBtn").onclick=doSave;

const drop=$("drop"),input=$("file");
drop.onclick=()=>input.click();
drop.onkeydown=e=>{if(e.key==="Enter"||e.key===" "){e.preventDefault();input.click()}};
input.onchange=()=>{if(input.files.length)handleFile([...input.files]);input.value=""};
["dragenter","dragover"].forEach(t=>drop.addEventListener(t,e=>{e.preventDefault();drop.classList.add("over")}));
["dragleave","drop"].forEach(t=>drop.addEventListener(t,e=>{e.preventDefault();drop.classList.remove("over")}));
drop.addEventListener("drop",e=>{if(e.dataTransfer.files.length)handleFile([...e.dataTransfer.files])});

// ---- matn ----
const tin=$("tin"),tout=$("tout"),tag=$("sampleTag"),tsp=$("tspell");
const textFixes=new Map();let spT=null;
function live(){
  let o=convert(tin.value,newCounts(),skipList());
  delete textFixes._re;o=UzSpell.applyFixes(o,textFixes);tout.value=o;
  clearTimeout(spT);spT=setTimeout(liveSpell,300);
}
function liveSpell(){
  tsp.textContent="";
  if(!speller||!spellOn()||!tout.value.trim())return;
  const all=findErrors(UzSpell.collectWords(tout.value,new Map()));
  const found=all.filter(f=>f.sugs.length),unk=all.filter(f=>!f.sugs.length);
  if(!found.length&&!unk.length){tsp.append(el("span","okmsg note","✓ Imlo xatosi topilmadi"));return}
  if(found.length){
    tsp.append(el("span","note","Imlo:"));
    for(const f of found){
      const b=el("button","fixb"+(f.confident?" sure":""));b.title=f.sugs.join(", ");
      b.append(el("s","",f.word),document.createTextNode(" → "),el("b","",f.sugs[0]));
      b.onclick=()=>{textFixes.set(f.word,f.sugs[0]);live()};tsp.append(b);
    }
    const sure=found.filter(f=>f.confident);
    if(sure.length>1){const a=el("button","mini","Aniq xatolarni tuzatish ("+sure.length+")");a.onclick=()=>{sure.forEach(f=>textFixes.set(f.word,f.sugs[0]));live()};tsp.append(a)}
  }
  if(unk.length)tsp.append(el("span","note","Lug‘atda yo‘q: "+unk.map(f=>f.word).join(", ")));
}
tin.addEventListener("input",()=>{tag.hidden=true;live()});
$("skip").addEventListener("input",live);$("cyr").addEventListener("change",live);$("spellOn").addEventListener("change",live);
$("clear").onclick=()=>{tin.value="";textFixes.clear();tag.hidden=true;live();tin.focus()};
$("copy").onclick=()=>{
  const m=$("copyMsg");
  navigator.clipboard.writeText(tout.value).then(()=>{m.textContent="Nusxa olindi."}).catch(()=>{tout.select();m.textContent="Matn belgilandi — Ctrl+C bosing."});
};
spellBadge();live();
})();
