import { useEffect, useState } from 'react';

export default function App() {
  const [data, setData] = useState(null);
  const [error, setError] = useState(null);

  async function refresh() {
    setError(null);
    try {
      const res = await fetch('/api/hello');
      if (!res.ok) {
        setError(`backend returned ${res.status}`);
        setData(null);
        return;
      }
      setData(await res.json());
    } catch (err) {
      setError(String(err));
      setData(null);
    }
  }

  useEffect(() => {
    refresh();
  }, []);

  return (
    <main>
      <header>
        <h1>Galley fullstack test</h1>
        <p className="muted">
          frontend (nginx + Vite) → backend (Go) → db (Postgres). Click below
          to bump a counter through the whole stack.
        </p>
      </header>

      <button type="button" onClick={refresh}>
        Bump counter
      </button>

      {error ? (
        <pre className="error">error: {error}</pre>
      ) : data ? (
        <pre className="data">{JSON.stringify(data, null, 2)}</pre>
      ) : (
        <p className="muted">loading…</p>
      )}
    </main>
  );
}
