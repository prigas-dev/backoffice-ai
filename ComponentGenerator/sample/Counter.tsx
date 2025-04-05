import { useState } from "react";

export function Component() {
  const [count, setCount] = useState(0);

  return (
    <div className="container">
      <div className="row">
        <div className="col">{count}</div>
        <div className="col">
          <button
            className="btn btn-danger"
            onClick={() => setCount(count + 1)}
          >
            +
          </button>
        </div>
      </div>
    </div>
  );
}
