export default {
  // Path to feature files
  paths: ['../tests/features/**/*.feature'],

  // Path to step definitions and support files
  import: [
    '../tests/features/step_definitions/**/*.ts',
    '../tests/features/support/**/*.ts'
  ],

  // Formatters
  format: [
    'progress-bar',
    'html:cucumber-report.html',
    'json:cucumber-report.json'
  ],

  // Parallel execution (adjust based on your CI capabilities)
  parallel: 1,

  // Tags to filter scenarios
  tags: process.env.CUCUMBER_TAGS || '',

  // Timeouts
  defaultTimeout: parseInt(process.env.DEFAULT_TIMEOUT || '30000'),

  // Publish report to Cucumber Studio (optional)
  publishQuiet: true,

  // Dry run (check step definitions without executing)
  dryRun: process.env.DRY_RUN === 'true',
};
