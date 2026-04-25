import { useEffect, useRef, useState } from 'react';

export default function App() {
  const [data, setData] = useState(null);
  const [error, setError] = useState(null);
  const [loading, setLoading] = useState(true);
  const [auto, setAuto] = useState(false);
  const [latencyMs, setLatencyMs] = useState(null);
  const [showRaw, setShowRaw] = useState(false);
  const timerRef = useRef(null);

  async function refresh() {
    setLoading(true);
    setError(null);
    const start = performance.now();
    try {
      const res = await fetch('/api/hello');
      if (!res.ok) {
        setError(`backend returned ${res.status}`);
        setData(null);
        return;
      }
      setData(await res.json());
      setLatencyMs(Math.round(performance.now() - start));
    } catch (err) {
      setError(String(err));
      setData(null);
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    refresh();
  }, []);

  // Auto-refresh: 2s tick when on, cleared when off. Uses a ref so we don't
  // chase setInterval through deps and accidentally double-up.
  useEffect(() => {
    if (!auto) {
      if (timerRef.current) {
        clearInterval(timerRef.current);
        timerRef.current = null;
      }
      return undefined;
    }
    timerRef.current = setInterval(refresh, 2000);
    return () => {
      if (timerRef.current) clearInterval(timerRef.current);
      timerRef.current = null;
    };
  }, [auto]);

  const visits = data?.visits ?? null;
  const cached = data?.cached === true;
  const hostname = data?.hostname ?? null;

  return (
    <main>
      <header>
        <p className="eyebrow">galley · fullstack preview</p>
        <h1>Hello from preview.</h1>
        <p className="lede">
          frontend (nginx) → backend (Go) → cache (Redis) → db (Postgres).
          Bumping the counter writes through Postgres and warms a 5s Redis
          memo. Repeated clicks inside that window come back cached.
        </p>
      </header>

      <div className="controls">
        <button type="button" className="primary" onClick={refresh} disabled={loading && !auto}>
          {loading && !auto ? 'Bumping…' : 'Bump counter'}
        </button>
        <label className="toggle">
          <input
            type="checkbox"
            checked={auto}
            onChange={(event) => setAuto(event.target.checked)}
          />
          <span>Auto-refresh · 2s</span>
        </label>
        {latencyMs !== null ? (
          <span className="latency" title="Round-trip from this tab to the backend">
            {latencyMs} ms
          </span>
        ) : null}
      </div>

      {error ? (
        <div className="error-card" role="alert">
          <strong>error</strong>
          <span>{error}</span>
        </div>
      ) : (
        <section className="stats" aria-label="Latest response">
          <div className="card">
            <p className="label">Visits</p>
            <p className="value tabular">{visits ?? '—'}</p>
          </div>
          <div className="card">
            <p className="label">Source</p>
            <p className="value">
              <span className={`pill ${cached ? 'pill-cached' : 'pill-fresh'}`}>
                {cached ? 'cache hit' : 'database'}
              </span>
            </p>
          </div>
          <div className="card">
            <p className="label">Container</p>
            <code className="value mono">
              {hostname ? hostname.slice(0, 12) : '—'}
            </code>
          </div>
        </section>
      )}

      <details className="raw" open={showRaw} onToggle={(event) => setShowRaw(event.target.open)}>
        <summary>Raw response</summary>
        <pre className="data">{data ? JSON.stringify(data, null, 2) : 'no data yet'}</pre>
      </details>
    </main>
  );
}
