export default {
  paths: ['tests/features/**/*.feature'],
  import: [
    'tests/features/step_definitions/**/*.ts',
    'tests/features/support/**/*.ts'
  ],
  format: ['progress'],
  parallel: 1,
  tags: process.env.CUCUMBER_TAGS || '',
  defaultTimeout: parseInt(process.env.DEFAULT_TIMEOUT || '60000'),
  publishQuiet: true,
};
