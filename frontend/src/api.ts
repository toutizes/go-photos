import { token } from './auth';
import { sortedPhotos, type Photo } from './model';
export async function request(path: string, signal?: AbortSignal): Promise<Response> {
  for (let attempt = 0; attempt < 2; attempt++) {
    const jwt = await token(attempt === 1);
    if (!jwt) throw new Error('Connectez-vous pour accéder aux photos.');
    const response = await fetch(path, { headers: { Authorization: `Bearer ${jwt}` }, signal, cache: 'no-store' });
    if (response.status === 401 && attempt === 0) continue;
    if (!response.ok) throw new Error(response.status === 403 ? 'Accès réservé au propriétaire.' : response.status === 401 ? 'Votre session a expiré. Reconnectez-vous.' : `Chargement impossible (${response.status}). Réessayez.`);
    return response;
  }
  throw new Error('Votre session a expiré. Reconnectez-vous.');
}
export async function json<T>(path: string, signal?: AbortSignal): Promise<T> { return (await request(path, signal)).json(); }
const searches = new Map<string, Photo[]>();
export const cachedSearch = (q: string) => searches.get(q);
export function clearSearches() { searches.clear(); }
// Always contact /db/q for an explicit search, even when rendering cached results.
// The existing backend records search history only on that endpoint.
export async function search(q: string, signal?: AbortSignal) {
  const photos = sortedPhotos(await json<Photo[]>(`/db/q?q=${encodeURIComponent(q)}`, signal), q);
  searches.delete(q); searches.set(q, photos);
  while (searches.size > 3) searches.delete(searches.keys().next().value!);
  return photos;
}
export async function download(path: string, name: string) {
  const response = await request(path);
  const blob = await response.blob();
  if (!blob.size) throw new Error('Le téléchargement est vide.');
  const url = URL.createObjectURL(blob);
  const link = document.createElement('a'); link.href = url; link.download = name;
  document.body.append(link); link.click(); link.remove();
  setTimeout(() => URL.revokeObjectURL(url), 60_000);
}
