module.exports = {
  ci: {
    collect: {
      // Number of runs to average
      numberOfRuns: 3,
      // Start server before testing
      startServerCommand: 'npm run preview',
      startServerReadyPattern: 'Local:',
      startServerReadyTimeout: 60000,
      // URLs to test
      url: [
        'http://localhost:4173/',
        'http://localhost:4173/graph/3d'
      ],
      // Chrome flags for consistent results
      chromeFlags: '--no-sandbox --disable-gpu --disable-dev-shm-usage',
      // Settings
      settings: {
        preset: 'desktop',
        throttling: {
          // Simulate fast 4G
          rttMs: 40,
          throughputKbps: 10240,
          cpuSlowdownMultiplier: 1
        }
      }
    },
    assert: {
      // Performance assertions
      assertions: {
        'categories:performance': ['warn', { minScore: 0.7 }],
        'categories:accessibility': ['warn', { minScore: 0.8 }],
        'categories:best-practices': ['warn', { minScore: 0.8 }],
        'categories:seo': ['warn', { minScore: 0.8 }],
        
        // Core Web Vitals
        'first-contentful-paint': ['warn', { maxNumericValue: 2000 }],
        'largest-contentful-paint': ['warn', { maxNumericValue: 2500 }],
        'interactive': ['warn', { maxNumericValue: 3500 }],
        'cumulative-layout-shift': ['warn', { maxNumericValue: 0.1 }],
        'total-blocking-time': ['warn', { maxNumericValue: 200 }],
        
        // Resource budgets
        'resource-summary:document:size': ['warn', { maxNumericValue: 50000 }],
        'resource-summary:script:size': ['warn', { maxNumericValue: 500000 }],
        'resource-summary:image:size': ['warn', { maxNumericValue: 1000000 }]
      }
    },
    upload: {
      // Upload to temporary storage
      target: 'temporary-public-storage'
    }
  }
};
