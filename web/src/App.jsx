import { useMemo, useState } from "react";

const sampleGoals = {
  "sample.log": "analyze errors and warnings in this log",
  "sales.csv": "summarize sales by category and region",
};

const samples = ["sample.log", "sales.csv"];

function App() {
  const [inputFile, setInputFile] = useState("sample.log");
  const [goal, setGoal] = useState(sampleGoals["sample.log"]);
  const [result, setResult] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const latestStep = useMemo(() => {
    if (!result?.steps?.length) return null;
    return result.steps[result.steps.length - 1];
  }, [result]);

  async function runAgent(event) {
    event.preventDefault();
    setLoading(true);
    setError("");
    setResult(null);
    try {
      const response = await fetch("/api/runs", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ goal, inputFile }),
      });
      const data = await response.json();
      setResult(data);
      if (!response.ok) {
        setError(data.error || data.errorMessage || data.error || data.status || "Run failed");
      }
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }

  function changeSample(nextFile) {
    setInputFile(nextFile);
    setGoal(sampleGoals[nextFile]);
  }

  return (
    <main className="app-shell">
      <section className="intro">
        <div>
          <p className="eyebrow">Code-as-action demo</p>
          <h1>CodeAct Agent</h1>
          <p className="intro-copy">
            A data analyst that turns a goal into executable Go code, runs it, and reports the result.
          </p>
        </div>
        <div className="status-strip">
          <Status label="Model" value="OpenAI" />
          <Status label="Runtime" value="Go" />
          <Status label="Mode" value="Web + CLI" />
        </div>
      </section>

      <section className="workspace">
        <form className="control-panel" onSubmit={runAgent}>
          <div className="field">
            <label htmlFor="sample">Sample file</label>
            <select id="sample" value={inputFile} onChange={(event) => changeSample(event.target.value)}>
              {samples.map((sample) => (
                <option key={sample} value={sample}>
                  {sample}
                </option>
              ))}
            </select>
          </div>

          <div className="field">
            <label htmlFor="goal">Goal</label>
            <textarea id="goal" value={goal} onChange={(event) => setGoal(event.target.value)} rows={5} />
          </div>

          <button className="run-button" type="submit" disabled={loading}>
            <svg aria-hidden="true" viewBox="0 0 24 24">
              <path d="M8 5v14l11-7z" />
            </svg>
            {loading ? "Running..." : "Run agent"}
          </button>

          {error && <p className="error-text">{error}</p>}

          <div className="explain">
            <h2>Flow</h2>
            <ol>
              <li>Goal becomes prompt.</li>
              <li>Model writes Go action.</li>
              <li>Backend runs action.</li>
              <li>Report is saved.</li>
            </ol>
          </div>
        </form>

        <div className="result-panel">
          <ResultHeader result={result} loading={loading} />
          <FlowGrid result={result} latestStep={latestStep} goal={goal} />
        </div>
      </section>
    </main>
  );
}

function Status({ label, value }) {
  return (
    <div>
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  );
}

function ResultHeader({ result, loading }) {
  if (loading) {
    return (
      <div className="result-header">
        <p className="eyebrow">Run status</p>
        <h2>Generating action</h2>
      </div>
    );
  }
  if (!result) {
    return (
      <div className="result-header">
        <p className="eyebrow">Run status</p>
        <h2>Ready</h2>
      </div>
    );
  }
  return (
    <div className="result-header">
      <p className="eyebrow">Run {result.id}</p>
      <h2>{result.status}</h2>
    </div>
  );
}

function FlowGrid({ result, latestStep, goal }) {
  return (
    <div className="flow-grid">
      <Panel title="1. User goal" content={result?.goal || goal} />
      <Panel title="2. Prompt sent to model" content={latestStep?.prompt || "Run the agent to see the prompt."} pre />
      <Panel title="3. Generated Go action" content={latestStep?.actionCode || "Generated code appears here."} pre />
      <Panel title="4. Execution output" content={latestStep?.output || latestStep?.error || "Execution output appears here."} pre />
      <Panel title="5. Final report" content={result?.reportText || result?.error || "Final report appears here."} pre wide />
    </div>
  );
}

function Panel({ title, content, pre = false, wide = false }) {
  return (
    <article className={wide ? "panel panel-wide" : "panel"}>
      <h3>{title}</h3>
      {pre ? <pre>{content}</pre> : <p>{content}</p>}
    </article>
  );
}

export default App;
