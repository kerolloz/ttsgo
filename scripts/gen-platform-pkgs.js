const fs = require('fs');
const path = require('path');

const apps = ['ttsgo', 'nego'];
const platforms = [
  { os: 'darwin', arch: 'arm64', npmOs: 'darwin', npmArch: 'arm64' },
  { os: 'darwin', arch: 'x64',   npmOs: 'darwin', npmArch: 'x64' },
  { os: 'linux',  arch: 'arm64', npmOs: 'linux',  npmArch: 'arm64' },
  { os: 'linux',  arch: 'x64',   npmOs: 'linux',  npmArch: 'x64' },
  { os: 'win32',  arch: 'arm64', npmOs: 'win32',  npmArch: 'arm64' },
  { os: 'win32',  arch: 'x64',   npmOs: 'win32',  npmArch: 'x64' },
];

apps.forEach(app => {
  platforms.forEach(p => {
    const pkgName = `@${app}/core-${p.npmOs}-${p.npmArch}`;
    const dir = path.join(__dirname, '..', 'packages', `${app}-${p.npmOs}-${p.npmArch}`);
    
    if (!fs.existsSync(dir)) {
      fs.mkdirSync(dir, { recursive: true });
    }
    
    const pkgJson = {
      name: pkgName,
      version: "0.1.0",
      description: `${app} binary for ${p.npmOs} ${p.npmArch}`,
      os: [p.os],
      cpu: [p.arch],
      license: "MIT",
      bin: {
        [app]: p.os === 'win32' ? `bin/${app}.exe` : `bin/${app}`
      },
      files: ["bin/"]
    };
    
    fs.writeFileSync(
      path.join(dir, 'package.json'),
      JSON.stringify(pkgJson, null, 2) + '\n'
    );
  });
});

console.log('Generated 12 platform package.json files.');
