import { App } from "@archivist/app.tsx";
import { makeDeps } from "@archivist/deps.ts";
import { render } from "preact";
import "./app.css";

const appRoot = document.createElement("div");

document.body.appendChild(appRoot);

render(<App deps={makeDeps()} />, appRoot);
