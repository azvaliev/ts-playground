const globals = require("globals");
const tseslint = require("typescript-eslint");

/** @type {import('eslint').Linter.Config[]} */
module.exports = tseslint.config([
  {
    files: ["**/*.{js,mjs,cjs,ts}"]
  },
  {
    files: ["**/*.js", "**/*.cjs"],
    languageOptions: {
      sourceType: "commonjs"
    }
  },
  {
    "files": ["**/*.ts"],
    languageOptions: {
      parser: '@typescript-eslint/parser',
      parserOptions: {
        sourceType: 'module'
      },
      globals: {
        ...globals.node,
        ...globals.es2020,
      }
    },
    plugins: {
      '@typescript-eslint': tseslint.plugin,
    },
    rules: {
      '@typescript-eslint/no-duplicate-enum-values': 'error',
      '@typescript-eslint/no-empty-object-type': 'error',
        '@typescript-eslint/no-misused-new': 'error',
    }
  },
]);
