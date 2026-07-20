package main

// reportCSS intentionally mirrors the blue Statistics page while remaining
// responsive on screen and print-friendly for ISP or regulator submissions.
const reportCSS = `:root{color-scheme:light dark}
*{box-sizing:border-box}
body{font:14px/1.5 "Segoe UI",Arial,sans-serif;margin:0;background:#111820;color:#eaf0f7}
.wrap{max-width:1180px;margin:auto;padding:28px}
.hero{background:linear-gradient(135deg,#0b6cff,#38bdf8);padding:28px;border-radius:18px;color:white;box-shadow:0 18px 50px #0005}
.hero-row{display:flex;align-items:flex-start;justify-content:space-between;gap:20px}
.hero h1{font-size:32px;line-height:1.15;margin:0 0 8px}
.hero p{margin:0;opacity:.92}
.print-button{appearance:none;border:1px solid #ffffff70;border-radius:12px;padding:10px 16px;background:#ffffff1f;color:white;font:600 14px "Segoe UI",Arial,sans-serif;cursor:pointer;white-space:nowrap}
.print-button:hover{background:#ffffff32}
.summary-grid{display:grid;grid-template-columns:repeat(3,minmax(0,1fr));gap:14px;margin-top:18px}
.metric{background:#18222d;border:1px solid #2b3948;border-radius:16px;padding:16px;min-height:98px}
.metric span{display:block;color:#8ec5ff;font-weight:600;margin-bottom:7px}
.metric strong{display:block;font-size:20px;line-height:1.3}
.metric small{display:block;margin-top:5px;opacity:.75}
.card{background:#18222d;border:1px solid #2b3948;border-radius:16px;padding:20px;margin-top:18px;overflow:auto}
.card h2{margin:0 0 12px}
.table-wrap{overflow:auto}
table{width:100%;border-collapse:collapse;min-width:760px}
th,td{padding:11px 12px;text-align:left;border-bottom:1px solid #2b3948;vertical-align:top}
th{color:#8ec5ff;font-weight:700}
tbody tr:hover{background:#ffffff08}
.note{background:linear-gradient(135deg,#0b6cff22,#38bdf822);border:1px solid #258cff66;border-left:5px solid #258cff;border-radius:14px;padding:15px 16px;margin-top:18px}
.muted{opacity:.78}
@media(max-width:760px){.wrap{padding:16px}.hero-row{display:block}.print-button{margin-top:16px}.summary-grid{grid-template-columns:1fr}.hero h1{font-size:27px}}
@media(prefers-color-scheme:light){body{background:#f3f6fa;color:#18222d}.metric,.card{background:white;border-color:#d9e2ec}.metric span,th{color:#1769d2}th,td{border-color:#e3e9ef}tbody tr:hover{background:#0b6cff08}.note{background:linear-gradient(135deg,#e9f3ff,#effbff);border-color:#8ec5ff}}
@media print{body{background:white;color:#111}.wrap{max-width:none;padding:0}.hero{box-shadow:none;background:#0b6cff!important;-webkit-print-color-adjust:exact;print-color-adjust:exact}.print-button{display:none}.metric,.card,.note{break-inside:avoid;background:white;color:#111;border-color:#ccd6e0}.metric span,th{color:#0b5ec2}.summary-grid{grid-template-columns:repeat(3,1fr)}table{min-width:0}th,td{font-size:11px;padding:7px}}`
