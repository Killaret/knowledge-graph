export default {
  paths: ['frontend/tests/features/**/*.feature'],
  require: [
    'frontend/tests/features/support/world.ts',
    'frontend/tests/features/support/hooks.ts',
    'frontend/tests/features/step_definitions/**/*.ts'
  ],
  import: ['tsx'],
  format: ['progress'],
  publishQuiet: true
};
