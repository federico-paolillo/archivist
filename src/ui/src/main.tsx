import { render } from "preact";
import { App } from "@archivist/app.tsx";

const appRoot = document.createElement("div");

document.body.appendChild(appRoot);

render(<App />, appRoot);
