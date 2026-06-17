import svelte from "rollup-plugin-svelte";
import commonjs from "@rollup/plugin-commonjs";
import resolve from "@rollup/plugin-node-resolve";
import livereload from "rollup-plugin-livereload";
import { terser } from "rollup-plugin-terser";
import copy from "rollup-plugin-copy";
import postcss from "rollup-plugin-postcss";
import autoprefixer from "autoprefixer";

const production = !process.env.ROLLUP_WATCH;

function serve() {
  let server;

  function toExit() {
    if (server) server.kill(0);
  }

  return {
    writeBundle() {
      if (server) return;
      server = require("child_process").spawn(
        "npm",
        ["run", "start", "--", "--dev"],
        {
          stdio: ["ignore", "inherit", "inherit"],
          shell: true,
        },
      );

      process.on("SIGTERM", toExit);
      process.on("exit", toExit);
    },
  };
}

export default {
  input: "src/main.js",
  output: {
    sourcemap: !production,
    format: "iife",
    name: "app",
    file: "dist/bundle.js",
  },
  plugins: [
    svelte({
      compilerOptions: {
        // enable run-time checks when not in production
        dev: !production,
      },
    }),
    postcss({
      minimize: true,
      extract: "bundle.css",
      sourceMap: !production,
      plugins: [autoprefixer()],
    }),
    // If you have external dependencies installed from
    // npm, you'll most likely need these plugins. In
    // some cases you'll need additional configuration -
    // consult the documentation for details:
    // https://github.com/rollup/plugins/tree/master/packages/commonjs
    resolve({
      browser: true,
      dedupe: ["svelte"],
    }),
    commonjs(),
    copy({
      targets: [
        { src: "src/index.html", dest: "dist/" },
        { src: "src/global.css", dest: "dist/" },
        { src: "src/assets", dest: "dist/" },
        { src: "node_modules/material-icons/iconfont/material-icons.css", dest: "dist/iconfont" },
        { src: "node_modules/material-icons/iconfont/material-icons.woff", dest: "dist/iconfont" },
        { src: "node_modules/material-icons/iconfont/material-icons.woff2", dest: "dist/iconfont" },
        //{ src: 'src/cssreset.css', dest: 'dist/' },
      ],
    }),

    // In dev mode, call `npm run start` once
    // the bundle has been generated
    !production && serve(),

    // Watch the `dist` directory and refresh the
    // browser on changes when not in production
    !production && livereload("dist"),

    // If we're building for production (npm run build
    // instead of npm run dev), minify
    production &&
      terser({
        compress: {
          drop_console: true,
        },
        format: {
          comments: false,
        },
      }),
  ],
  watch: {
    clearScreen: false,
  },
};
