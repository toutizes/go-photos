import { useEffect, useRef, useState } from 'react';
import { readRoute, routeHash, type Route } from './model';
export function useRoute() {
  const [route,setRoute] = useState(readRoute);
  const closing = useRef<Route | null>(null);
  useEffect(() => { const read = () => { if(closing.current) { const next=closing.current; closing.current=null; history.replaceState({level: next.id === undefined ? undefined : 'photo'},'',routeHash(next)); setRoute(next); } else setRoute(readRoute()); }; window.addEventListener('popstate',read); window.addEventListener('hashchange',read); return () => {window.removeEventListener('popstate',read);window.removeEventListener('hashchange',read);}; },[]);
  function navigate(next: Route, replace = false, level?: 'photo' | 'full') {
    history[replace ? 'replaceState' : 'pushState'](level ? {level} : {},'',routeHash(next)); setRoute(next);
  }
  function close() {
    if ((route.full && history.state?.level === 'full') || (!route.full && history.state?.level === 'photo')) { closing.current = {...route,full:false,id:route.full ? route.id : undefined}; history.back(); }
    else navigate({...route,full:false,id:route.full ? route.id : undefined},true,route.full ? 'photo' : undefined);
  }
  return {route,navigate,close};
}
