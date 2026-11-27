import {defineConfig} from '@playwright/test';
import {fileURLToPath} from 'node:url';
export default defineConfig({
 testDir:'.',testMatch:'integration.spec.mjs',fullyParallel:false,workers:1,timeout:30000,
 reporter:[['line'],['json',{outputFile:process.env.TB2_BROWSER_REPORT || 'test-results/integration-report.json'}]],
 use:{baseURL:process.env.TB2_FIXTURE_URL || 'http://127.0.0.1:8098',browserName:'chromium',channel:process.env.PLAYWRIGHT_CHANNEL || undefined,headless:true,viewport:{width:1280,height:900},screenshot:'only-on-failure'},
 webServer:process.env.TB2_FIXTURE_URL ? undefined : {command:'go run internal/ui/tests/integration_browser.go',cwd:fileURLToPath(new URL('../../../../',import.meta.url)),url:'http://127.0.0.1:8098',reuseExistingServer:false,timeout:120000},
});

