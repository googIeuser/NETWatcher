import type { ReactNode } from 'react'
import {icons} from './Icons'
import type {PageId} from '../types'
import type {TranslationKey} from '../lib/i18n'
const items:{id:PageId;label:TranslationKey;icon:ReactNode}[]=[{id:'dashboard',label:'dashboard',icon:icons.dashboard},{id:'statistics',label:'statistics',icon:icons.statistics},{id:'outages',label:'outages',icon:icons.outages},{id:'reports',label:'reports',icon:icons.reports},{id:'targets',label:'targets',icon:icons.targets},{id:'settings',label:'settings',icon:icons.settings}]
export function Sidebar({page,onChange,t,version}:{page:PageId;onChange(v:PageId):void;t:(k:TranslationKey)=>string;version:string}){return <aside className="sidebar"><div className="brand"><img src="/app-icon.png"/><div><strong>NetWatcher</strong><span>{t('connectionIntelligence')}</span></div></div><nav>{items.map(item=><button className={page===item.id?'nav-item active':'nav-item'} onClick={()=>onChange(item.id)} key={item.id}>{item.icon}<span>{t(item.label)}</span></button>)}</nav><div className="sidebar-footer"><span className="version-pill">{version||'3.0 preview'}</span><small>Go · Wails · React</small></div></aside>}
