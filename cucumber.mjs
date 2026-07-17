import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = dirname(fileURLToPath(import.meta.url));
const featureRoot = join(__dirname, 'frontend', 'tests', 'features');

export default {
  paths: [join(featureRoot, '**', '*.feature')],
  require: [
    join(featureRoot, 'support', 'world.ts'),
    join(featureRoot, 'support', 'hooks.ts'),
    join(featureRoot, 'step_definitions', '**', '*.ts')
  ],
  import: ['tsx'],
  format: ['progress'],
  publishQuiet: true
};
