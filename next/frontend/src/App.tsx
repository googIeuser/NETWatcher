import {useCallback,useEffect,useMemo,useState} from 'react'
import {Sidebar} from './components/Sidebar'
import {Dashboard} from './pages/Dashboard'
import {StatisticsPage} from './pages/StatisticsPage'
import {OutagesPage} from './pages/OutagesPage'
import {ReportsPage} from './pages/ReportsPage'
import {TargetsPage} from './pages/TargetsPage'
import {SettingsPage} from './pages/SettingsPage'
import {bridge} from './lib/bridge'
import {translator} from './lib/i18n'
import type {PageId,Settings,Snapshot} from './types'

export default function App(){
  const [page,setPage]=useState<PageId>('dashboard')
  const [snapshot,setSnapshot]=useState<Snapshot|null>(null)
  const [settings,setSettings]=useState<Settings|null>(null)
  const [busy,setBusy]=useState(false)
  const [toast,setToast]=useState<{kind:'error'|'success';text:string}|null>(null)
  const language=settings?.language??'en';const t=useMemo(()=>translator(language),[language])
  const notify=(kind:'error'|'success',text:string)=>{setToast({kind,text});window.setTimeout(()=>setToast(null),3600)}
  const run=useCallback(async<T,>(action:()=>Promise<T>):Promise<T>=>{try{return await action()}catch(error){const text=error instanceof Error?error.message:String(error);notify('error',text);throw error}},[])
  useEffect(()=>{let live=true;void Promise.all([bridge.getSnapshot(),bridge.getSettings()]).then(([snap,cfg])=>{if(live){setSnapshot(snap);setSettings(cfg)}}).catch(e=>notify('error',String(e)));return()=>{live=false}},[])
  useEffect(()=>{if(!settings)return;document.documentElement.dataset.theme=settings.theme;document.documentElement.lang=settings.language},[settings])
  useEffect(()=>{const id=window.setInterval(()=>void bridge.getSnapshot().then(setSnapshot).catch(()=>{}),1000);return()=>window.clearInterval(id)},[])
  const toggle=async()=>{if(!snapshot)return;setBusy(true);try{setSnapshot(await run(()=>snapshot.monitoring?bridge.stopMonitoring():bridge.startMonitoring()))}finally{setBusy(false)}}
  const saveSettings=async(v:Settings)=>{const next=await run(()=>bridge.saveSettings(v));setSettings(next);notify('success',t('saved'))}
  const add=async(v:string)=>setSettings(await run(()=>bridge.addTarget(v)))
  const edit=async(a:string,b:string)=>setSettings(await run(()=>bridge.editTarget(a,b)))
  const remove=async(v:string)=>setSettings(await run(()=>bridge.removeTarget(v)))
  const loadStats=useCallback((hours:number)=>run(()=>bridge.getStatistics(hours)),[run])
  const loadOutages=useCallback((days:number)=>run(()=>bridge.getOutages(days)),[run])
  if(!snapshot||!settings)return <div className="boot-screen"><img src="/app-icon.png"/><strong>NetWatcher</strong><span>{t('loading')}</span></div>
  return <div className="app-shell"><Sidebar page={page} onChange={setPage} t={t} version={snapshot.version}/><main>{page==='dashboard'&&<Dashboard snapshot={snapshot} settings={settings} busy={busy} onToggle={toggle} t={t}/>} {page==='statistics'&&<StatisticsPage load={loadStats} t={t}/>} {page==='outages'&&<OutagesPage load={loadOutages} t={t}/>} {page==='reports'&&<ReportsPage html={h=>run(()=>bridge.generateHTMLReport(h))} evidence={d=>run(()=>bridge.generateEvidenceReport(d))} exportZip={h=>run(()=>bridge.exportDiagnostics(h))} openLogs={()=>run(()=>bridge.openLogs())} t={t}/>} {page==='targets'&&<TargetsPage snapshot={snapshot} settings={settings} onAdd={add} onEdit={edit} onRemove={remove} t={t}/>} {page==='settings'&&<SettingsPage value={settings} onSave={saveSettings} onCheckUpdate={()=>run(()=>bridge.checkUpdates())} onOpenRelease={url=>run(()=>bridge.openRelease(url))} t={t}/>}</main>{toast&&<div className={`toast ${toast.kind}`}>{toast.text}</div>}</div>
}
