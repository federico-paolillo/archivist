import { render } from "preact";
import "./index.css";
import { App } from "./app.tsx";

const appRoot = document.createElement("div");

document.body.appendChild(appRoot);

render(<App />, appRoot);
