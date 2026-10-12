import { useEffect, useRef, useState } from "react";
import {
  Images,
  LayoutGrid,
  History,
  Search,
  LogOut,
  Shield,
  HelpCircle,
} from "lucide-react";
import { watchAccount, login, logout, isAdmin, type Account } from "./auth";
import { cachedSearch, clearSearches, search } from "./api";
import { clearImages } from "./images";
import { todayQuery, type Photo, type Route } from "./model";
import { useRoute } from "./useRoute";
import { Grid } from "./Grid";
import { Viewer } from "./Viewer";
import { ActivityPage, AlbumsPage, AdminPage } from "./Collections";
import { Downloads } from "./Downloads";
export function App() {
  const [account, setAccount] = useState<Account | null | undefined>(undefined),
    [authError, setAuthError] = useState("");
  const { route, navigate, close } = useRoute();
  const [draft, setDraft] = useState(route.q),
    [help, setHelp] = useState(false),
    [searchVersion, setSearchVersion] = useState(0);
  useEffect(
    () =>
      watchAccount((a) => {
        clearSearches();
        clearImages();
        setAccount(a);
      }),
    [],
  );
  useEffect(() => setDraft(route.q), [route.q]);
  const goSearch = (q: string) => {
    setSearchVersion((v) => v + 1);
    navigate({ page: "photos", q: q.trim() || todayQuery(), full: false });
  };
  async function signIn() {
    try {
      setAuthError("");
      await login();
    } catch {
      setAuthError(
        "Connexion interrompue. Autorisez la fenêtre Google puis réessayez.",
      );
    }
  }
  if (account === undefined)
    return (
      <div className="login">
        <p>Chargement…</p>
      </div>
    );
  if (!account)
    return (
      <main className="login">
        <img src={`${import.meta.env.BASE_URL}logo-transparent.png`} alt="" />
        <p className="eyebrow">LES PHOTOS, AU FIL DU TEMPS</p>
        <h1>Bienvenue sur Toutizes</h1>
        <p>Des moments à retrouver et à partager.</p>
        <button className="primary" onClick={signIn}>
          Se connecter avec Google
        </button>
        <p className="hint">
          Identifiez-vous pour accéder aux photos.
          <br />
          Votre identité et vos recherches sont visibles par le propriétaire.
        </p>
        {authError && <p role="alert">{authError}</p>}
      </main>
    );
  const tabs = [
    { page: "activity", label: "Récemment", Icon: History },
    { page: "albums", label: "Albums", Icon: Images },
    { page: "photos", label: "Photos", Icon: LayoutGrid },
    ...(isAdmin(account)
      ? [{ page: "admin", label: "Admin", Icon: Shield }]
      : []),
  ] as const;
  return (
    <div key={account.email} className={`app ${route.full ? "is-full" : ""}`}>
      {!route.full && (
        <header className="app-header">
          <a
            className="brand"
            href="#/activity"
            aria-label="Toutizes — Récemment"
          >
            <img
              src={`${import.meta.env.BASE_URL}logo-transparent.png`}
              alt=""
            />
          </a>
          <nav aria-label="Navigation principale">
            {tabs.map(({ page, label, Icon }) => (
              <button
                key={page}
                aria-label={label}
                aria-current={route.page === page ? "page" : undefined}
                className={route.page === page ? "active" : ""}
                onClick={() =>
                  navigate({
                    page: page as Route["page"],
                    q: page === "photos" ? todayQuery() : "",
                    full: false,
                  })
                }
              >
                <Icon size={18} />
                <span>{label}</span>
              </button>
            ))}
          </nav>
          <form
            className="search-form"
            onSubmit={(e) => {
              e.preventDefault();
              if (route.page === "albums") {
                setSearchVersion((v) => v + 1);
                navigate({ page: "albums", q: draft.trim(), full: false });
              } else goSearch(draft);
            }}
          >
            <Search size={18} />
            <input
              aria-label={
                route.page === "albums"
                  ? "Rechercher des albums"
                  : "Rechercher des photos"
              }
              placeholder="Un lieu, un prénom, une année…"
              value={draft}
              onChange={(e) => setDraft(e.target.value)}
            />
            <button type="submit" aria-label="Lancer la recherche">
              ↵
            </button>
          </form>
          <button
            className="icon-button help"
            aria-label="Aide à la recherche"
            onClick={() => setHelp(!help)}
          >
            <HelpCircle size={20} />
          </button>
          <button
            className="icon-button signout"
            aria-label="Se déconnecter"
            title={account.email ?? ""}
            onClick={() => {
              void logout().catch(() =>
                setAuthError("Déconnexion impossible. Réessayez."),
              );
            }}
          >
            <LogOut size={19} />
          </button>
        </header>
      )}
      {help && !route.full && (
        <section className="search-help">
          <strong>Retrouver une photo</strong>
          <p>
            Séparez les mots-clés par des virgules : <b>famille, vacances</b>.
            Essayez une année (<b>2020</b>), une date (<b>2020-07-14</b>) ou un
            intervalle d’années (<b>2020--2022</b>). Les pastilles sous les
            photos lancent une nouvelle recherche.
          </p>
          <p>
            Utilisez « <b>{"titre:<texte>"}</b> » pour ne chercher que dans les
            albums qui contiennent {"<texte>"} dans leur nom.
          </p>
          <button onClick={() => setHelp(false)}>Fermer l’aide</button>
        </section>
      )}
      {authError && <p role="alert">{authError}</p>}
      <main className="content">
        {route.page === "photos" ? (
          <Photos
            searchVersion={searchVersion}
            route={{ ...route, q: route.q || todayQuery() }}
            navigate={navigate}
            close={close}
            onSearch={goSearch}
          />
        ) : route.page === "activity" ? (
          <ActivityPage
            onSearch={goSearch}
            onPhoto={(q, id) =>
              navigate({ page: "photos", q, id, full: false })
            }
          />
        ) : route.page === "albums" ? (
          <AlbumsPage
            key={route.q}
            query={route.q}
            searchVersion={searchVersion}
            onSearch={goSearch}
          />
        ) : isAdmin(account) ? (
          <AdminPage onSearch={goSearch} />
        ) : (
          <section className="empty">
            <h1>Accès réservé au propriétaire</h1>
          </section>
        )}
      </main>
    </div>
  );
}
function Photos({
  route,
  navigate,
  close,
  onSearch,
  searchVersion,
}: {
  searchVersion: number;
  route: Route;
  navigate: (r: Route, replace?: boolean, level?: "photo" | "full") => void;
  close: () => void;
  onSearch: (q: string) => void;
}) {
  const [result, setResult] = useState<{ q: string; photos: Photo[] }>({
    q: route.q,
    photos: cachedSearch(route.q) ?? [],
  });
  const [loading, setLoading] = useState(true),
    [error, setError] = useState(""),
    [refresh, setRefresh] = useState(0);
  useEffect(() => {
    const controller = new AbortController();
    setLoading(true);
    setError("");
    setResult({ q: route.q, photos: cachedSearch(route.q) ?? [] });
    search(route.q, controller.signal).then(
      (photos) => {
        setResult({ q: route.q, photos });
        setLoading(false);
      },
      (e) => {
        if (!controller.signal.aborted) {
          setError(e.message);
          setLoading(false);
        }
      },
    );
    return () => controller.abort();
  }, [route.q, refresh, searchVersion]);
  const photos = result.q === route.q ? result.photos : [];
  const lastPhoto = useRef<number | undefined>(undefined);
  if (route.id !== undefined) lastPhoto.current = route.id;
  return (
    <div className="photos-page">
      {route.id === undefined && (
        <div className="page-heading">
          <div>
            <h1>Les photos</h1>
          </div>
          <div className="heading-actions">
            <Downloads query={route.q} count={photos.length} />
          </div>
        </div>
      )}
      {error && (
        <div className="error" role="alert">
          {error}
          <button onClick={() => setRefresh((n) => n + 1)}>Réessayer</button>
        </div>
      )}
      {!photos.length ? (
        <section className="empty">
          {route.id !== undefined && (
            <button onClick={close}>Revenir aux photos</button>
          )}
          <h2>
            {loading
              ? "Chargement des photos…"
              : route.q
                ? `Aucune photo trouvée pour « ${route.q} »`
                : "Aucune photo trouvée"}
          </h2>
          {!loading && <p>Essayez un autre mot-clé ou une année.</p>}
        </section>
      ) : (
        <>
          <Grid
            key={route.q}
            activeId={lastPhoto.current}
            photos={photos}
            hidden={route.id !== undefined}
            open={(p) =>
              navigate({ ...route, id: p.Id, full: false }, false, "photo")
            }
          />
          {route.id !== undefined && (
            <Viewer
              photos={photos}
              id={route.id}
              full={route.full}
              onStep={(id) =>
                navigate({ ...route, id }, true, history.state?.level)
              }
              onFull={() => navigate({ ...route, full: true }, false, "full")}
              onClose={close}
              onSearch={onSearch}
            />
          )}
        </>
      )}
    </div>
  );
}
