package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"ais-demo/internal/ais"
)

var secret []byte
var auditLog []string
var auditNotify = make(chan struct{}, 1)

func main() {
	secret = []byte(os.Getenv("AIS_SECRET"))
	if len(secret) == 0 {
		secret = []byte("dev-secret-change-me")
	}

    // Model status endpoints and UI
    http.HandleFunc("/model/status", handleModelStatus)
    http.HandleFunc("/model/pull", handleModelPull)
    http.HandleFunc("/model/list", handleModelList)
    http.HandleFunc("/model/select", handleModelSelect)
    http.HandleFunc("/audit/stream", handleAuditStream)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, "/chat", http.StatusFound) })
    http.HandleFunc("/chat", handleChat)
	http.HandleFunc("/confirm", handleConfirm)
	http.HandleFunc("/execute", handleExecute)
    http.HandleFunc("/api/chat/send", handleChatSend)
    http.HandleFunc("/api/chat/plan", handlePlan)

	log.Println("AIS demo on http://localhost:8890")
	log.Fatal(http.ListenAndServe(":8890", nil))
}

var tmpl = template.Must(template.New("index").Parse(`<!doctype html><html><body>
<h2>AIS MVP (Ollama)</h2>
<div id="modelStatus"></div>
<div id="progress" style="width:400px;height:20px;border:1px solid #888;display:none"><div id="bar" style="height:100%;background:#4caf50;width:0%"></div></div>
<form method="POST" action="/confirm">
Intent (what and why):<br/>
<textarea name="purpose" rows=4 cols=60>Summarize this text for an executive brief</textarea><br/>
Data Classes (csv): <input name="dataclasses" value="internal,derived"/><br/>
Prompt: <br/>
<textarea name="prompt" rows=6 cols=60>Provide a concise summary of our quarterly goals and results.</textarea><br/>
<button type="submit">Confirm Intent</button>
</form>
</br>
<script>
async function checkModel(){
  const s = document.getElementById('modelStatus');
  const p = document.getElementById('progress');
  const b = document.getElementById('bar');
  try{
    const r = await fetch('/model/status');
    const j = await r.json();
    if(j.present){ s.textContent = 'Model present: '+j.model; return }
    s.textContent = 'Model missing: '+j.model+' — pulling...'; p.style.display='block';
    const resp = await fetch('/model/pull');
    const reader = resp.body.getReader();
    const decoder = new TextDecoder();
    let buf = '';
    while(true){
      const {done, value} = await reader.read(); if(done) break;
      buf += decoder.decode(value, {stream:true});
      let idx;
      while((idx = buf.indexOf('\n')) >= 0){
        const line = buf.slice(0, idx).trim(); buf = buf.slice(idx+1);
        if(!line) continue; try{
          const o = JSON.parse(line);
          if(o.percent){ b.style.width = o.percent+'%'; }
        }catch(e){}
      }
    }
    s.textContent = 'Model ready'; b.style.width='100%';
  }catch(e){ s.textContent = 'Model status error: '+e }
}
checkModel();
</script>
</body></html>`))

var confirmTmpl = template.Must(template.New("confirm").Parse(`<!doctype html><html><body>
<h3>Confirm Intent</h3>
<pre>{{.UIAJSON}}</pre>
<h3>Planned Steps</h3>
<pre>{{.APAJSON}}</pre>
<form method="POST" action="/execute">
<input type="hidden" name="uia" value='{{.UIA}}'/>
<input type="hidden" name="apa" value='{{.APA}}'/>
<input type="hidden" name="apr" value='{{.APr}}'/>
<input type="hidden" name="prompt" value='{{.Prompt}}'/>
<button type="submit">Execute with Enforcement</button>
</form>
</body></html>`))

func handleIndex(w http.ResponseWriter, r *http.Request) {
	_ = tmpl.Execute(w, nil)
}

var chatTmpl = template.Must(template.New("chat").Parse(`<!doctype html><html>
<head>
<meta charset="utf-8"/>
<meta name="viewport" content="width=device-width,initial-scale=1"/>
<title>AIS Chat</title>
<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/codemirror@5.65.16/lib/codemirror.css">
<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/codemirror@5.65.16/theme/darcula.css">
<script src="https://cdn.jsdelivr.net/npm/codemirror@5.65.16/lib/codemirror.js"></script>
<script src="https://cdn.jsdelivr.net/npm/codemirror@5.65.16/mode/javascript/javascript.js"></script>
<style>
body{margin:0;background:#0f1115;color:#e6e6e6;font-family:Inter,system-ui,Arial,sans-serif}
.container{display:flex;flex-direction:column;height:100vh}
.topbar{display:flex;gap:12px;align-items:center;padding:12px 16px;background:#0b0d11;border-bottom:1px solid #1b1f27}
.status{font-size:12px;color:#9aa4b2}
.progress{width:280px;height:8px;border:1px solid #2a2f3a;border-radius:6px;display:none}
.bar{height:100%;background:#4caf50;width:0%;border-radius:6px}
.main{flex:1;display:flex}
.sidebar{width:320px;border-right:1px solid #1b1f27;padding:16px;display:none}
.sidebar h3{margin:0 0 8px 0;font-size:14px;color:#9aa4b2}
.sidebar .codebox{background:#0b0d11;border:1px solid #1b1f27;border-radius:8px;max-height:40vh;overflow:hidden}
.CodeMirror{height:160px}
.chat{flex:1;display:flex;flex-direction:column}
.messages{flex:1;overflow:auto;padding:16px;display:flex;flex-direction:column;gap:12px}
.msg{max-width:70%;padding:12px 14px;border-radius:12px;line-height:1.4}
.user{align-self:flex-end;background:#1b2330;border:1px solid #2a3545}
.assistant{align-self:flex-start;background:#141824;border:1px solid #212835}
.composer{display:flex;gap:8px;padding:12px 16px;border-top:1px solid #1b1f27;background:#0b0d11}
.composer textarea{flex:1;background:#0f131a;color:#e6e6e6;border:1px solid #1f2733;border-radius:10px;padding:12px;min-height:48px;max-height:140px}
.button{background:#2563eb;color:#fff;border:none;border-radius:10px;padding:12px 16px;cursor:pointer}
.iconbtn{background:transparent;border:1px solid #2a2f3a;color:#9aa4b2;border-radius:8px;padding:8px 10px;cursor:pointer}
.pill{font-size:12px;padding:2px 8px;border-radius:999px;background:#121620;border:1px solid #1b2130;color:#9aa4b2}
.modal{position:fixed;inset:0;background:rgba(0,0,0,0.5);display:none;align-items:center;justify-content:center}
.dialog{background:#0b0d11;border:1px solid #1b1f27;padding:16px 18px;border-radius:12px;width:520px}
.dialog h3{margin:0 0 8px 0}
.row{display:flex;gap:8px;align-items:center}
.overlay{position:fixed;inset:0;background:rgba(3,6,12,0.78);display:none;align-items:center;justify-content:center;z-index:20}
.loadcard{background:#0b0d11;border:1px solid #1b1f27;border-radius:12px;padding:20px 22px;min-width:520px;display:flex;flex-direction:column;gap:12px;align-items:center}
.bigprogress{width:420px;height:10px;border:1px solid #2a2f3a;border-radius:8px}
.bigbar{height:100%;background:#4caf50;width:0%;border-radius:8px}
.disabled{opacity:.6;pointer-events:none;filter:grayscale(20%)}
</style>
</head>
<body>
<div class="container">
  <div class="topbar">
    <button id="intentToggle" class="iconbtn">⫶ Intent</button>
    <span class="status" id="modelStatus"></span>
    <span class="pill">AIS Demo</span>
    <select id="modelSel" class="iconbtn" style="margin-left:auto"></select>
    <button id="auditToggle" class="iconbtn">Audit</button>
  </div>
  <div class="main">
    <div class="sidebar" id="intentPanel">
      <h3>Intent (UIA)</h3>
      <div class="codebox"><textarea id="uiaEditor"></textarea></div>
      <h3>Plan (APA)</h3>
      <div class="codebox"><textarea id="apaEditor"></textarea></div>
      <h3>Alignment (APr)</h3>
      <div class="codebox"><textarea id="aprEditor"></textarea></div>
      <h3>Agent Steps</h3>
      <div class="codebox"><textarea id="stepsEditor"></textarea></div>
    </div>
    <div class="chat">
      <div class="messages" id="messages"></div>
      <div class="composer">
        <textarea id="input" placeholder="Message... (Ctrl+Enter to send)"></textarea>
        <button id="send" class="button">Send</button>
        <button id="runAgent" class="iconbtn">Run Agent</button>
        <button id="replan" class="iconbtn">Re‑Plan</button>
      </div>
    </div>
    <div class="sidebar" id="auditPanel" style="display:none">
      <h3>Audit (live)</h3>
      <div class="codebox"><textarea id="auditBox"></textarea></div>
    </div>
  </div>
</div>

<div class="overlay" id="loadingOverlay">
  <div class="loadcard">
    <h3 style="margin:0">Preparing model…</h3>
    <div class="bigprogress"><div class="bigbar" id="bigbar"></div></div>
    <div class="status" id="loadText">Downloading model…</div>
  </div>
</div>
<div class="modal" id="confirmModal">
  <div class="dialog">
    <h3>Confirm Intent</h3>
    <p>Please confirm the intent before proceeding.</p>
    <pre id="confirmIntent"></pre>
    <div class="row" style="justify-content:flex-end">
      <button id="cancelBtn" class="iconbtn">Cancel</button>
      <button id="okBtn" class="button">Confirm</button>
    </div>
  </div>
  </div>

<script>
const msgs = [];
let intentConfirmed = false;
let currentIntent = null;
let uiaEd, apaEd, aprEd, stepsEd;
let auditEd;

function addMsg(role, content){
  msgs.push({role, content});
  const m = document.getElementById('messages');
  const div = document.createElement('div');
  div.className = 'msg '+(role==='user'?'user':'assistant');
  div.textContent = content;
  m.appendChild(div);
  m.scrollTop = m.scrollHeight;
}

function initEditors(){
  uiaEd = CodeMirror.fromTextArea(document.getElementById('uiaEditor'), {mode:'application/json', theme:'darcula', lineNumbers:false, readOnly:false});
  apaEd = CodeMirror.fromTextArea(document.getElementById('apaEditor'), {mode:'application/json', theme:'darcula', lineNumbers:false, readOnly:true});
  aprEd = CodeMirror.fromTextArea(document.getElementById('aprEditor'), {mode:'application/json', theme:'darcula', lineNumbers:false, readOnly:true});
  stepsEd = CodeMirror.fromTextArea(document.getElementById('stepsEditor'), {mode:'application/json', theme:'darcula', lineNumbers:false, readOnly:false});
  auditEd = CodeMirror.fromTextArea(document.getElementById('auditBox'), {mode:'application/json', theme:'darcula', lineNumbers:false, readOnly:true});
  stepsEd.setValue(JSON.stringify([
    {tool:'ollama.generate', prompt:'Summarize the conversation so far.'},
    {tool:'ollama.generate', prompt:'Extract 3 key bullet points.'},
    {tool:'ollama.generate', prompt:'Produce a concise executive summary.'}
  ], null, 2));
}

function showIntent(uia, apa, apr){
  if(uiaEd){ uiaEd.setValue(uia ? JSON.stringify(uia, null, 2) : ''); }
  if(apaEd){
    if(typeof apa === 'string'){ apaEd.setValue(apa); }
    else { apaEd.setValue(apa ? JSON.stringify(apa, null, 2) : ''); }
  }
  if(aprEd){
    if(typeof apr === 'string'){ aprEd.setValue(apr); }
    else { aprEd.setValue(apr ? JSON.stringify(apr, null, 2) : ''); }
  }
}

function setUIEnabled(on){
  const chat = document.querySelector('.chat');
  const composer = document.querySelector('.composer');
  const input = document.getElementById('input');
  const send = document.getElementById('send');
  if(on){
    chat.classList.remove('disabled');
    composer.classList.remove('disabled');
    input.disabled = false;
    send.disabled = false;
  } else {
    chat.classList.add('disabled');
    composer.classList.add('disabled');
    input.disabled = true;
    send.disabled = true;
  }
}

async function checkModel(){
  const s = document.getElementById('modelStatus');
  const overlay = document.getElementById('loadingOverlay');
  const bigbar = document.getElementById('bigbar');
  await refreshModelList();
  setUIEnabled(false);
  try{
    const r = await fetch('/model/status');
    const j = await r.json();
    if(j.present){ s.textContent = 'Model ready: '+j.model; overlay.style.display='none'; setUIEnabled(true); return }
    overlay.style.display='flex'; s.textContent = 'Model downloading…';
    const resp = await fetch('/model/pull');
    const reader = resp.body.getReader();
    const decoder = new TextDecoder();
    let buf='';
    while(true){
      const {done,value} = await reader.read(); if(done) break;
      buf += decoder.decode(value,{stream:true});
      let idx; while((idx=buf.indexOf('\n'))>=0){
        const line = buf.slice(0,idx).trim(); buf = buf.slice(idx+1);
        if(!line) continue;
        try{ const o = JSON.parse(line); if(o.percent){ bigbar.style.width = o.percent+'%'; } }catch(e){}
      }
    }
    // After stream ends, poll until present
    for(let i=0;i<30;i++){
      const rr = await fetch('/model/status');
      const jj = await rr.json();
      if(jj.present){ s.textContent='Model ready'; overlay.style.display='none'; setUIEnabled(true); return }
      await new Promise(res=>setTimeout(res,1000));
    }
    s.textContent='Model not ready yet. Please wait…';
  }catch(e){ s.textContent = 'Model status error: '+e }
}

document.getElementById('intentToggle').onclick = ()=>{
  const p = document.getElementById('intentPanel');
  p.style.display = (p.style.display==='block'?'none':'block');
}

function needConfirmIntent(prompt){
  const risky = ['email','send','post','export','delete'];
  return risky.some(k => prompt.toLowerCase().includes(k));
}

async function send(){
  const ta = document.getElementById('input');
  const text = ta.value.trim();
  if(!text) return;
  if(document.getElementById('loadingOverlay').style.display==='flex'){ return }
  addMsg('user', text);
  ta.value = '';

  const purpose = 'Chat: '+text.slice(0,120);
  const uia = { "@type":"UIA", id:'urn:uia:'+Date.now(), subject:{id:'user:demo'}, purpose, constraints:{dataClasses:['internal','derived'], timeWindow:{notAfter:new Date(Date.now()+10*60*1000).toISOString()}}, riskBudget:{level:1, maxWrites:0, maxRecords:1000}, policyProfile:'chat-readonly', proof:{} };
  currentIntent = uia;
  showIntent(uia, 'Planning…', 'Pending…');
  if(!intentConfirmed && needConfirmIntent(text)){
    document.getElementById('confirmIntent').textContent = JSON.stringify(uia,null,2);
    const modal = document.getElementById('confirmModal'); modal.style.display='flex';
    return;
  }

  const body = { messages: msgs, uia: currentIntent };
  const r = await fetch('/api/chat/send', { method:'POST', headers:{'content-type':'application/json'}, body: JSON.stringify(body)});
  if(!r.ok){ addMsg('assistant', 'Error: '+await r.text()); return }
  const j = await r.json();
  showIntent(j.uia, j.apa, j.apr);
  addMsg('assistant', j.assistant);
}

async function runAgent(){
  if(document.getElementById('loadingOverlay').style.display==='flex'){ return }
  // Simulate a 3-step agent plan; each step goes through intent→APA/APr→IBE guard
  let steps = [];
  try { steps = JSON.parse(stepsEd.getValue()); } catch(e) { alert('Invalid steps JSON'); return }
  let ctx = msgs.map(m => (m.role+': '+m.content)).join('\n');
  for(let i=0;i<steps.length;i++){
    const s = steps[i];
    const uia = { "@type":"UIA", id:'urn:uia:'+Date.now()+':'+i, subject:{id:'user:demo'}, purpose:'Agent step '+(i+1)+': '+(s.prompt||s.url||s.tool), constraints:{dataClasses:['internal','derived'], timeWindow:{notAfter:new Date(Date.now()+10*60*1000).toISOString()}}, riskBudget:{level:1, maxWrites:0, maxRecords:1000, maxExternalCalls: (s.tool==='http.get'?1:0)}, policyProfile:'agent-readonly', proof:{} };
    showIntent(uia, {info:'planning step '+(i+1)+'…'}, {info:'pending…'});
    let r;
    if(s.tool==='http.get'){
      r = await fetch('/api/chat/send', { method:'POST', headers:{'content-type':'application/json'}, body: JSON.stringify({messages:[{role:'user', content:'Fetch URL: '+s.url}], uia, tool:'http.get', url:s.url})});
    } else {
      const body = { messages: [{role:'user', content: ctx+'\n\nTask: '+s.prompt}], uia };
      r = await fetch('/api/chat/send', { method:'POST', headers:{'content-type':'application/json'}, body: JSON.stringify(body)});
    }
    if(!r.ok){ addMsg('assistant', 'Step '+(i+1)+' error: '+await r.text()); return }
    const j = await r.json();
    showIntent(j.uia, j.apa, j.apr);
    addMsg('assistant', '(step '+(i+1)+') '+j.assistant);
    ctx += '\nAssistant: '+j.assistant;
  }
}

async function replan(){
  let uia;
  try { uia = JSON.parse(uiaEd.getValue()); } catch(e) { alert('Invalid UIA JSON'); return }
  const body = { uia };
  const r = await fetch('/api/chat/plan', { method:'POST', headers:{'content-type':'application/json'}, body: JSON.stringify(body)});
  if(!r.ok){ alert(await r.text()); return }
  const j = await r.json();
  showIntent(j.uia, j.apa, j.apr);
}

function toggleAudit(){
  const a = document.getElementById('auditPanel');
  a.style.display = (a.style.display==='block'?'none':'block');
}

function startAudit(){
  const es = new EventSource('/audit/stream');
  es.onmessage = (e)=>{
    try { const obj = JSON.parse(e.data); const prev = auditEd.getValue(); auditEd.setValue((prev?prev+'\n':'')+JSON.stringify(obj)); } catch(_) {}
  };
}

document.getElementById('send').onclick = send;
document.getElementById('input').addEventListener('keydown', (e)=>{ if(e.key==='Enter' && e.ctrlKey){ send(); }});
document.getElementById('runAgent').onclick = runAgent;
document.getElementById('replan').onclick = replan;
document.getElementById('auditToggle').onclick = toggleAudit;

document.getElementById('okBtn').onclick = ()=>{ intentConfirmed = true; document.getElementById('confirmModal').style.display='none'; send(); };
document.getElementById('cancelBtn').onclick = ()=>{ document.getElementById('confirmModal').style.display='none'; };

initEditors();
checkModel();
startAudit();
</script>
</body></html>`))

func handleChat(w http.ResponseWriter, r *http.Request) {
	_ = chatTmpl.Execute(w, nil)
}

func handleConfirm(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "post only", 405)
		return
	}
	purpose := r.FormValue("purpose")
    classes := r.FormValue("dataclasses")
	prompt := r.FormValue("prompt")

    dc := splitCSV(classes)
    // Ensure 'derived' is permitted for generated content in this demo
    if !contains(dc, "derived") { dc = append(dc, "derived") }

    uia := ais.UIA{
		Type: "UIA",
		ID:   "urn:uia:" + nowID(),
		Subject: ais.Principal{ID: "user:demo"},
		Purpose: purpose,
        Constraints: ais.Constraints{DataClasses: dc, TimeWindow: ais.TimeBound{NotAfter: time.Now().Add(10 * time.Minute)}},
		RiskBudget: ais.RiskBudget{Level: 1, MaxWrites: 0, MaxRecords: 1000},
		PolicyProfile: "research-readonly",
		Proof: map[string]any{},
	}
	apa := ais.APA{
		Type: "APA",
		ID:   nowID(),
		UIA:  uia.ID,
		Model: ais.ModelInfo{Hash: "ollama-local"},
		Steps: []ais.APAStep{{
			ID: "s1", Tool: "ollama.generate",
			Args: map[string]any{"prompt": prompt},
			Expected: ais.StepExpected{DataClasses: []string{"derived"}, Writes: 0},
			Alignment: ais.StepAlignment{Score: 0.9, Why: "generate summary"},
		}},
		Totals: ais.APATotals{PredictedWrites: 0, PredictedRecords: 1},
		Proof: map[string]any{},
	}
    apr := ais.APr{Type: "APr", ID: nowID(), UIA: uia.ID, APA: apa.ID, Method: "semantic-entailment-v1", Evidence: ais.APrEvidence{Coverage: 1.0, Risk: 0.0}, Proof: map[string]any{}}

	uiaJSON, _ := json.MarshalIndent(uia, "", "  ")
	apaJSON, _ := json.MarshalIndent(apa, "", "  ")
	aprJSON, _ := json.Marshal(apr)

	_ = confirmTmpl.Execute(w, map[string]any{
		"UIAJSON": string(uiaJSON),
		"APAJSON": string(apaJSON),
		"UIA":     string(uiaJSON),
		"APA":     string(apaJSON),
		"APr":     string(aprJSON),
		"Prompt":  prompt,
	})
}

func handleExecute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "post only", 405)
		return
	}
	var uia ais.UIA
	var apa ais.APA
	var apr ais.APr
	_ = json.Unmarshal([]byte(r.FormValue("uia")), &uia)
	_ = json.Unmarshal([]byte(r.FormValue("apa")), &apa)
	_ = json.Unmarshal([]byte(r.FormValue("apr")), &apr)
	prompt := r.FormValue("prompt")

	tca := ais.TCA{ID: "urn:tca:ollama.generate@1", Operations: []ais.TCAOperation{{Name: "ollama.generate", Effects: ais.OperationEffects{Writes: 0, DataClasses: []string{"derived"}}}}, Operator: "local"}

    ibe := ais.IBE{Type: "IBE", ID: "urn:ibe:" + nowID(), UIARef: uia.ID, APAStepRef: "s1", APrRef: apr.ID, TCARef: tca.ID, Nonce: nowID(), Exp: time.Now().Add(2 * time.Minute)}
    ibeForSig := ibe
    ibeForSig.Sig = ""
    sig, _ := ais.SignJWSObject(secret, ibeForSig)
    ibe.Sig = sig

	if err := ais.VerifyIBE(ais.GuardConfig{Secret: secret, MinAlignment: 0.8}, ibe, apr, uia, apa, tca); err != nil {
		http.Error(w, "blocked by guard: "+err.Error(), 403)
		return
	}

	client := &ais.OllamaClient{BaseURL: envDefault("OLLAMA_URL", "http://localhost:11434"), Model: envDefault("OLLAMA_MODEL", "llama3"), HTTP: http.DefaultClient}
    // If model is still missing (e.g., first-time), trigger pull via UI route
    if ok, _ := client.HasModel(); !ok {
        http.Error(w, "model not present yet; please wait for pull to complete", 409)
        return
    }
	resp, err := client.Generate(prompt)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
    // audit event for legacy execute path
    ev := map[string]any{"ts": time.Now().UTC().Format(time.RFC3339), "uia": uia.ID, "apa": apa.ID, "ibe": ibe.ID, "tool": "ollama.generate", "ok": true}
    b, _ := json.Marshal(ev)
    auditLog = append(auditLog, string(b))
    select { case auditNotify <- struct{}{}: default: }
	w.Header().Set("content-type", "text/plain")
	_, _ = w.Write([]byte(resp))
}

func nowID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func splitCSV(s string) []string {
	var out []string
	cur := ""
	for i := 0; i < len(s); i++ {
		if s[i] == ',' {
			if cur != "" { out = append(out, trimSpace(cur)) }
			cur = ""
			continue
		}
		cur += string(s[i])
	}
	if cur != "" { out = append(out, trimSpace(cur)) }
	return out
}

func trimSpace(s string) string {
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\t') { s = s[1:] }
	for len(s) > 0 && (s[len(s)-1] == ' ' || s[len(s)-1] == '\t') { s = s[:len(s)-1] }
	return s
}

func envDefault(k, v string) string {
	if x := os.Getenv(k); x != "" { return x }
	return v
}

func contains(slice []string, item string) bool {
    for _, s := range slice {
        if s == item { return true }
    }
    return false
}

type chatMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
type chatReq struct {
	Messages []chatMsg   `json:"messages"`
	UIA      ais.UIA     `json:"uia"`
    Tool     string      `json:"tool"`
    URL      string      `json:"url"`
}
type chatResp struct {
	Assistant string   `json:"assistant"`
	UIA       ais.UIA   `json:"uia"`
	APA       ais.APA   `json:"apa"`
	APr       ais.APr   `json:"apr"`
}

func handleChatSend(w http.ResponseWriter, r *http.Request) {
	var req chatReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", 400)
		return
	}
    // Build APA depending on tool
    var sb strings.Builder
    for _, m := range req.Messages {
        if m.Role == "user" {
            _, _ = sb.WriteString("User: ")
            _, _ = sb.WriteString(m.Content)
            _, _ = sb.WriteString("\n")
        }
        if m.Role == "assistant" {
            _, _ = sb.WriteString("Assistant: ")
            _, _ = sb.WriteString(m.Content)
            _, _ = sb.WriteString("\n")
        }
    }
    prompt := sb.String()
    step := ais.APAStep{ID: "s1", Tool: "ollama.generate", Args: map[string]any{"prompt": prompt}, Expected: ais.StepExpected{DataClasses: []string{"derived"}, Writes: 0}, Alignment: ais.StepAlignment{Score: 1.0}}
    if req.Tool == "http.get" {
        step = ais.APAStep{ID: "s1", Tool: "http.get", Args: map[string]any{"url": req.URL}, Expected: ais.StepExpected{DataClasses: []string{"derived"}, Writes: 0}, Alignment: ais.StepAlignment{Score: 1.0}}
    }
    apa := ais.APA{Type: "APA", ID: nowID(), UIA: req.UIA.ID, Model: ais.ModelInfo{Hash: "ollama-local"}, Steps: []ais.APAStep{step}, Totals: ais.APATotals{PredictedWrites: 0, PredictedRecords: 1}, Proof: map[string]any{}}
	apr := ais.APr{Type: "APr", ID: nowID(), UIA: req.UIA.ID, APA: apa.ID, Method: "semantic-entailment-v1", Evidence: ais.APrEvidence{Coverage: 1.0, Risk: 0.0}, Proof: map[string]any{}}

    tca := ais.TCA{ID: "urn:tca:ollama.generate@1", Operations: []ais.TCAOperation{{Name: "ollama.generate", Effects: ais.OperationEffects{Writes: 0, DataClasses: []string{"derived"}}}}, Operator: "local"}
    if req.Tool == "http.get" {
        tca = ais.TCA{ID: "urn:tca:http.get@1", Operations: []ais.TCAOperation{{Name: "http.get", Effects: ais.OperationEffects{Writes: 0, DataClasses: []string{"derived"}}}}, Operator: "local"}
    }
	ibe := ais.IBE{Type: "IBE", ID: "urn:ibe:" + nowID(), UIARef: req.UIA.ID, APAStepRef: "s1", APrRef: apr.ID, TCARef: tca.ID, Nonce: nowID(), Exp: time.Now().Add(2 * time.Minute)}
	ibeForSig := ibe
	ibeForSig.Sig = ""
    sig, _ := ais.SignJWSObject(secret, ibeForSig)
	ibe.Sig = sig

	if err := ais.VerifyIBE(ais.GuardConfig{Secret: secret, MinAlignment: 0.8}, ibe, apr, req.UIA, apa, tca); err != nil {
		http.Error(w, "blocked by guard: "+err.Error(), 403)
		return
	}
    var respText string
    var err error
    if req.Tool == "http.get" {
        httpTool := &ais.HTTPTool{HTTP: http.DefaultClient}
        respText, err = httpTool.Get(req.URL)
    } else {
        client := &ais.OllamaClient{BaseURL: envDefault("OLLAMA_URL", "http://localhost:11434"), Model: envDefault("OLLAMA_MODEL", "llama3"), HTTP: http.DefaultClient}
        if ok, _ := client.HasModel(); !ok {
            http.Error(w, "model not present yet; please wait for pull to complete", 409)
            return
        }
        respText, err = client.Generate(prompt)
    }
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
    // audit event
    ev := map[string]any{"ts": time.Now().UTC().Format(time.RFC3339), "uia": req.UIA.ID, "apa": apa.ID, "ibe": ibe.ID, "tool": step.Tool, "ok": true}
    b, _ := json.Marshal(ev)
    auditLog = append(auditLog, string(b))
    select { case auditNotify <- struct{}{}: default: }
	w.Header().Set("content-type", "application/json")
	_ = json.NewEncoder(w).Encode(chatResp{Assistant: respText, UIA: req.UIA, APA: apa, APr: apr})
}

func handlePlan(w http.ResponseWriter, r *http.Request) {
    var req struct{ UIA ais.UIA `json:"uia"` }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "bad request", 400)
        return
    }
    // Build a plan (APA/APr) without executing tools
    step := ais.APAStep{ID: "s1", Tool: "ollama.generate", Args: map[string]any{"prompt": "planned"}, Expected: ais.StepExpected{DataClasses: []string{"derived"}, Writes: 0}, Alignment: ais.StepAlignment{Score: 1.0}}
    apa := ais.APA{Type: "APA", ID: nowID(), UIA: req.UIA.ID, Model: ais.ModelInfo{Hash: "ollama-local"}, Steps: []ais.APAStep{step}, Totals: ais.APATotals{PredictedWrites: 0, PredictedRecords: 1}, Proof: map[string]any{}}
    apr := ais.APr{Type: "APr", ID: nowID(), UIA: req.UIA.ID, APA: apa.ID, Method: "semantic-entailment-v1", Evidence: ais.APrEvidence{Coverage: 1.0, Risk: 0.0}, Proof: map[string]any{}}
    w.Header().Set("content-type", "application/json")
    _ = json.NewEncoder(w).Encode(map[string]any{"uia": req.UIA, "apa": apa, "apr": apr})
}


func handleModelStatus(w http.ResponseWriter, r *http.Request) {
    c := &ais.OllamaClient{BaseURL: envDefault("OLLAMA_URL", "http://localhost:11434"), Model: envDefault("OLLAMA_MODEL", "llama3"), HTTP: http.DefaultClient}
    ok, err := c.HasModel()
    type resp struct{ Present bool `json:"present"`; Model string `json:"model"`; Error string `json:"error,omitempty"` }
    out := resp{Present: ok, Model: c.Model}
    if err != nil { out.Error = err.Error() }
    w.Header().Set("content-type", "application/json")
    _ = json.NewEncoder(w).Encode(out)
}

func handleModelPull(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("content-type", "application/x-ndjson")
    w.Header().Set("cache-control", "no-cache")
    flusher, _ := w.(http.Flusher)
    c := &ais.OllamaClient{BaseURL: envDefault("OLLAMA_URL", "http://localhost:11434"), Model: envDefault("OLLAMA_MODEL", "llama3"), HTTP: http.DefaultClient}
    err := c.PullStream(func(m map[string]any) error {
        b, _ := json.Marshal(m)
        _, _ = w.Write(append(b, '\n'))
        if flusher != nil { flusher.Flush() }
        return nil
    })
    if err != nil {
        _, _ = io.WriteString(w, `{"error":"`+template.HTMLEscapeString(err.Error())+`"}\n`)
        if flusher != nil { flusher.Flush() }
    }
}

func handleModelList(w http.ResponseWriter, r *http.Request) {
    c := &ais.OllamaClient{BaseURL: envDefault("OLLAMA_URL", "http://localhost:11434"), Model: envDefault("OLLAMA_MODEL", "llama3"), HTTP: http.DefaultClient}
    models, err := c.ListModels()
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
    w.Header().Set("content-type", "application/json")
    _ = json.NewEncoder(w).Encode(map[string]any{"models": models})
}

func handleModelSelect(w http.ResponseWriter, r *http.Request) {
    var p struct{ Model string `json:"model"` }
    if err := json.NewDecoder(r.Body).Decode(&p); err != nil || p.Model == "" {
        http.Error(w, "bad request", 400)
        return
    }
    // Persist selection via env for this process (demo); in real app store in session
    _ = os.Setenv("OLLAMA_MODEL", p.Model)
    w.WriteHeader(204)
}

func handleAuditStream(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("content-type", "text/event-stream")
    w.Header().Set("cache-control", "no-cache")
    flusher, _ := w.(http.Flusher)

    // send existing entries
    for _, line := range auditLog {
        _, _ = io.WriteString(w, "data: "+line+"\n\n")
    }
    if flusher != nil { flusher.Flush() }

    notify := auditNotify
    for {
        <-notify
        if len(auditLog) == 0 { continue }
        last := auditLog[len(auditLog)-1]
        _, _ = io.WriteString(w, "data: "+last+"\n\n")
        if flusher != nil { flusher.Flush() }
    }
}



