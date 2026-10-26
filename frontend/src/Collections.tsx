import { useEffect, useRef, useState } from "react";
import { ArrowUpRight, CalendarDays, Search, FolderOpen } from "lucide-react";
import { json } from "./api";
import { useImage } from "./images";
import {
  dateLabel,
  thumbPath,
  todayQuery,
  type ActivityGroup,
  type Album,
  type QueryEntry,
} from "./model";
function useResource<T>(path: string, searchVersion = 0) {
  const [data, setData] = useState<T>(),
    [error, setError] = useState(""),
    [loading, setLoading] = useState(true),
    [version, setVersion] = useState(0);
  useEffect(() => {
    const c = new AbortController();
    setLoading(true);
    setError("");
    json<T>(path, c.signal).then(
      (value) => {
        setData(value);
        setLoading(false);
      },
      (e) => {
        if (!c.signal.aborted) {
          setError(e.message);
          setLoading(false);
        }
      },
    );
    return () => c.abort();
  }, [path, version, searchVersion]);
  return { data, error, loading, refresh: () => setVersion((n) => n + 1) };
}
function Status({
  loading,
  error,
  retry,
}: {
  loading: boolean;
  error: string;
  retry: () => void;
}) {
  return error ? (
    <div className="error" role="alert">
      {error}
      <button onClick={retry}>Réessayer</button>
    </div>
  ) : loading ? (
    <p className="loading collection-loading" role="status">
      Chargement…
    </p>
  ) : null;
}
function useNearScreen() {
  const ref = useRef<HTMLDivElement>(null),
    [visible, setVisible] = useState(false);
  useEffect(() => {
    const observer = new IntersectionObserver(
      ([entry]) => setVisible(entry.isIntersecting),
      { rootMargin: "250px" },
    );
    if (ref.current) observer.observe(ref.current);
    return () => observer.disconnect();
  }, []);
  return { ref, visible };
}
function PreviewTile({
  id,
  name,
  visible,
  open,
}: {
  id: number;
  name: string;
  visible: boolean;
  open: (id: number) => void;
}) {
  const image = useImage(visible ? thumbPath(id) : undefined);
  return (
    <button
      onClick={() => open(id)}
      aria-label={`Ouvrir ${name || "la photo"}`}
    >
      {image.url ? (
        <img src={image.url} alt="" draggable={false} />
      ) : image.error ? (
        <span className="image-error">Photo indisponible</span>
      ) : (
        <span className="image-placeholder" />
      )}
    </button>
  );
}
function PreviewStrip({
  ids,
  names,
  open,
}: {
  ids: number[];
  names: string[];
  open: (id: number) => void;
}) {
  const { ref, visible } = useNearScreen();
  return (
    <div
      className="preview-strip"
      ref={ref}
      style={{
        gridTemplateColumns: `repeat(${Math.max(3, ids.length)},minmax(0,1fr))`,
      }}
    >
      {ids.map((id, i) => (
        <PreviewTile
          key={id}
          id={id}
          name={names[i]}
          visible={visible}
          open={open}
        />
      ))}
    </div>
  );
}
export function ActivityPage({
  onSearch,
  onPhoto,
}: {
  onSearch: (q: string) => void;
  onPhoto: (q: string, id: number) => void;
}) {
  const { data, error, loading, refresh } = useResource<{
    groups: ActivityGroup[];
  }>("/db/recent-keyword-groups");
  return (
    <div className="collection-page scroll">
      <div className="page-heading">
        <div>
          <h1>Récemment</h1>
        </div>
        <button className="on-this-day" onClick={() => onSearch(todayQuery())}>
          <CalendarDays size={16} />
          Ce jour-là
          <ArrowUpRight size={16} />
        </button>
      </div>
      <Status loading={loading} error={error} retry={refresh} />
      <div className="activity-list">
        {data?.groups.map((group, index) => {
          const title = group.keywords[0]?.keyword ?? "Photos";
          return (
            <article className="activity-card" key={`${title}-${index}`}>
              <div className="activity-copy">
                {group.keywords.map((k) => (
                  <button key={k.keyword} onClick={() => onSearch(k.keyword)}>
                    <span>{k.keyword}</span>
                  </button>
                ))}
              </div>
              <PreviewStrip
                ids={group.recent_images.map((p) => p.id)}
                names={group.recent_images.map((p) => p.name)}
                open={(id) => onPhoto(title, id)}
              />
            </article>
          );
        })}
      </div>
      {!loading && !error && !data?.groups.length && (
        <section className="empty">
          <h2>Pas encore d’activité</h2>
          <p>Parcourez les albums ou recherchez une année.</p>
        </section>
      )}
    </div>
  );
}
function AlbumCard({
  album,
  onSearch,
}: {
  album: Album;
  onSearch: (q: string) => void;
}) {
  const { ref, visible } = useNearScreen();
  const cover = useImage(
    visible && album.Cov ? thumbPath(album.Cov) : undefined,
  );
  return (
    <div ref={ref} className="album-card">
      <button onClick={() => onSearch(`album:${album.Id}`)}>
        <div className="album-cover">
          {cover.url ? (
            <img src={cover.url} alt="" draggable={false} />
          ) : (
            <span className="image-placeholder" />
          )}
          <span className="album-count">{album.Nimgs} photos</span>
        </div>
        <div className="album-description">
          <span className="eyebrow">{dateLabel(album.Ats)}</span>
          <h2>{album.Id.split("/").at(-1)}</h2>
          <p>
            {album.Id}
            <ArrowUpRight size={16} />
          </p>
        </div>
      </button>
    </div>
  );
}
export function AlbumsPage({
  query,
  searchVersion,
  onSearch,
}: {
  query: string;
  searchVersion: number;
  onSearch: (q: string) => void;
}) {
  const { data, error, loading, refresh } = useResource<Album[]>(
    `/db/q?q=${encodeURIComponent(query || "albums:")}&kind=album`,
    searchVersion,
  );
  const [year, setYear] = useState("");
  const albums = (data ?? [])
    .filter((a) => a.Nimgs > 0)
    .sort((a, b) => b.Ats - a.Ats || a.Id.localeCompare(b.Id));
  const years = [
    ...new Set(
      albums.map((a) => new Date(a.Ats * 1000).getFullYear().toString()),
    ),
  ];
  const shown = albums.filter(
    (a) => !year || new Date(a.Ats * 1000).getFullYear().toString() === year,
  );
  return (
    <div className="collection-page scroll">
      <div className="page-heading">
        <div>
          <h1>Les albums</h1>
        </div>
        <select
          className="album-year-filter"
          aria-label="Année des albums"
          value={year}
          onChange={(e) => setYear(e.target.value)}
        >
          <option value="">Toutes les années</option>
          {years.map((y) => (
            <option key={y}>{y}</option>
          ))}
        </select>
      </div>
      <Status loading={loading} error={error} retry={refresh} />
      <div className="album-grid">
        {shown.map((a) => (
          <AlbumCard key={a.Id} album={a} onSearch={onSearch} />
        ))}
      </div>
      {!loading && !error && !shown.length && (
        <section className="empty">
          <FolderOpen size={32} />
          <h2>
            {query
              ? `Aucun album trouvé pour « ${query} »`
              : "Aucun album trouvé"}
          </h2>
          <p>Essayez un autre mot-clé ou une autre année.</p>
        </section>
      )}
    </div>
  );
}
export function AdminPage({ onSearch }: { onSearch: (q: string) => void }) {
  const { data, error, loading, refresh } = useResource<{
    users: Record<string, QueryEntry[] | null>;
  }>("/db/user-queries");
  const [filter, setFilter] = useState("");
  const users = Object.entries(data?.users ?? {})
    .filter(([email]) => email.toLowerCase().includes(filter.toLowerCase()))
    .sort((a, b) =>
      (b[1]?.[0]?.timestamp ?? "").localeCompare(a[1]?.[0]?.timestamp ?? ""),
    );
  return (
    <div className="collection-page scroll">
      <div className="page-heading">
        <div>
          <p className="eyebrow">ESPACE PROPRIÉTAIRE</p>
          <h1>Les visites</h1>
          <p>Les dernières recherches de chaque visiteur.</p>
        </div>
      </div>
      <div className="collection-filters">
        <label>
          <Search size={17} />
          <input
            aria-label="Filtrer les visiteurs"
            placeholder="Un visiteur…"
            value={filter}
            onChange={(e) => setFilter(e.target.value)}
          />
        </label>
      </div>
      <Status loading={loading} error={error} retry={refresh} />
      <div className="admin-list">
        {users.map(([email, queries]) => (
          <details className="admin-card" key={email}>
            <summary>{email}</summary>
            <table>
              <thead>
                <tr>
                  <th>Date</th>
                  <th>Recherche</th>
                  <th>Type</th>
                </tr>
              </thead>
              <tbody>
                {(queries ?? []).map((q, i) => (
                  <tr key={i}>
                    <td>{new Date(q.timestamp).toLocaleString("fr")}</td>
                    <td>
                      <button
                        className="text-button"
                        onClick={() => onSearch(q.query)}
                      >
                        {q.query}
                      </button>
                    </td>
                    <td>{q.kind === "album" ? "Albums" : "Photos"}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </details>
        ))}
      </div>
      {!loading && !error && !users.length && (
        <section className="empty">
          <h2>Aucune recherche enregistrée</h2>
        </section>
      )}
    </div>
  );
}
