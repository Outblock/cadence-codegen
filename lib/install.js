#!/usr/bin/env node

const fs = require('fs');
const path = require('path');
const https = require('https');
const { execSync } = require('child_process');
const tar = require('tar');

const BINARY_NAME = 'cadence-codegen';
const GITHUB_REPO = 'outblock/cadence-codegen';

function getPlatform() {
  const platform = process.platform;
  const arch = process.arch;
  
  // Map Node.js platform/arch to Go GOOS/GOARCH
  const platformMap = {
    'darwin': 'darwin',
    'linux': 'linux',
    'win32': 'windows'
  };
  
  const archMap = {
    'x64': 'amd64',
    'arm64': 'arm64'
  };
  
  const goos = platformMap[platform];
  const goarch = archMap[arch];
  
  if (!goos || !goarch) {
    throw new Error(`Unsupported platform: ${platform}-${arch}`);
  }
  
  return { goos, goarch };
}

function getArchiveUrl(version) {
  const { goos, goarch } = getPlatform();
  
  // Map to GoReleaser's naming convention
  const osMap = {
    'darwin': 'Darwin',
    'linux': 'Linux', 
    'windows': 'Windows'
  };
  
  const archMap = {
    'amd64': 'x86_64',
    'arm64': 'arm64'
  };
  
  const os = osMap[goos];
  const arch = archMap[goarch];
  const ext = goos === 'windows' ? 'zip' : 'tar.gz';
  
  const filename = `${BINARY_NAME}_${os}_${arch}.${ext}`;
  
  return `https://github.com/${GITHUB_REPO}/releases/download/v${version}/${filename}`;
}

function downloadAndExtract(url, dest) {
  return new Promise((resolve, reject) => {
    const { goos } = getPlatform();
    const tempDir = path.dirname(dest) + '/tmp';
    const archiveFile = path.join(tempDir, 'archive');
    
    // Create temp directory
    if (!fs.existsSync(tempDir)) {
      fs.mkdirSync(tempDir, { recursive: true });
    }
    
    const file = fs.createWriteStream(archiveFile);
    
    https.get(url, (response) => {
      if (response.statusCode === 302 || response.statusCode === 301) {
        return downloadAndExtract(response.headers.location, dest);
      }
      
      if (response.statusCode !== 200) {
        reject(new Error(`Failed to download: ${response.statusCode}`));
        return;
      }
      
      response.pipe(file);
      
      file.on('finish', async () => {
        file.close();
        
        try {
          if (goos === 'windows') {
            // For Windows, we'll need to handle ZIP files
            // For now, fallback to go install
            throw new Error('Windows binary extraction via ZIP not yet supported');
          } else {
            // Extract tar.gz
            await tar.x({
              file: archiveFile,
              cwd: tempDir,
              strip: 0
            });
            
            // Find the binary in extracted files
            const extractedBinary = path.join(tempDir, BINARY_NAME);
            if (fs.existsSync(extractedBinary)) {
              // Move binary to final destination
              fs.copyFileSync(extractedBinary, dest);
              fs.chmodSync(dest, 0o755);
            } else {
              throw new Error(`Binary ${BINARY_NAME} not found in extracted archive`);
            }
          }
          
          // Clean up temp directory
          fs.rmSync(tempDir, { recursive: true, force: true });
          resolve();
          
        } catch (error) {
          // Clean up on error
          fs.rmSync(tempDir, { recursive: true, force: true });
          reject(error);
        }
      });
      
      file.on('error', (error) => {
        fs.rmSync(tempDir, { recursive: true, force: true });
        reject(error);
      });
      
    }).on('error', (error) => {
      fs.rmSync(tempDir, { recursive: true, force: true });
      reject(error);
    });
  });
}

async function install() {
  try {
    const packageJson = JSON.parse(fs.readFileSync(path.join(__dirname, '../package.json'), 'utf8'));
    const version = packageJson.version;
    
    const binDir = path.join(__dirname, '../bin');
    const binaryPath = path.join(binDir, BINARY_NAME);
    
    // Create bin directory if it doesn't exist
    if (!fs.existsSync(binDir)) {
      fs.mkdirSync(binDir, { recursive: true });
    }
    
    const url = getArchiveUrl(version);
    console.log(`Downloading cadence-codegen archive from ${url}...`);
    
    await downloadAndExtract(url, binaryPath);
    console.log('✅ cadence-codegen binary installed successfully');
    
  } catch (error) {
    console.error('❌ Failed to install cadence-codegen binary:', error.message);
    console.log('Attempting to fall back to go install...');
    
    try {
      execSync(`go install github.com/${GITHUB_REPO}@latest`, { stdio: 'inherit' });
      console.log('✅ Installed via go install as fallback');
    } catch (goError) {
      console.error('❌ Go install fallback also failed:', goError.message);
      console.log('Please install manually:');
      console.log(`  brew install ${BINARY_NAME}`);
      console.log(`  or go install github.com/${GITHUB_REPO}@latest`);
      process.exit(1);
    }
  }
}

if (require.main === module) {
  install();
}