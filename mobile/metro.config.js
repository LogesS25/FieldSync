const { getDefaultConfig } = require('expo/metro-config');
const { withNativeWind } = require('nativewind/metro');
const path = require('path');

const config = getDefaultConfig(__dirname);

// zustand's package.json "exports" map only defines "react-native" (native)
// and "import" (ESM) conditions for its subpaths — no "browser" condition.
// Expo's web platform condition set is ['browser'], which doesn't match
// "react-native", so Metro falls through to "import" and resolves to
// zustand's esm/*.mjs build. That build uses `import.meta.env` (a
// Vite-style dead-code check in the devtools middleware) — valid only
// inside a real ES module, but Metro serves the web bundle as a classic
// script, so `import.meta` is a hard SyntaxError that crashes the entire
// bundle before any app code runs. Force zustand's CJS build (its "main"
// field, no import.meta) directly, bypassing exports-map resolution for
// this one package rather than disabling it project-wide.
const upstreamResolveRequest = config.resolver.resolveRequest;
config.resolver.resolveRequest = (context, moduleName, platform) => {
  if (moduleName === 'zustand' || moduleName.startsWith('zustand/')) {
    const subpath = moduleName === 'zustand' ? 'index.js' : `${moduleName.slice('zustand/'.length)}.js`;
    return {
      type: 'sourceFile',
      filePath: path.join(__dirname, 'node_modules', 'zustand', subpath),
    };
  }
  if (upstreamResolveRequest) {
    return upstreamResolveRequest(context, moduleName, platform);
  }
  return context.resolveRequest(context, moduleName, platform);
};

module.exports = withNativeWind(config, { input: './src/global.css' });
