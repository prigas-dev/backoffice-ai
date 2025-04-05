import { useState } from "react";

export function Component() {
  const [name, setName] = useState("");

  return (
    <div className="container">
      <div className="row">
        <div className="col">Hello {name}</div>
        <div className="col">
          <input
            type="text"
            name="name"
            id="name"
            value={name}
            onInput={(evt) => setName(evt.currentTarget.value)}
          />
        </div>
      </div>
    </div>
  );
}
