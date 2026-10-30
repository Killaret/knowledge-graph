import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = dirname(fileURLToPath(import.meta.url));
const featureRoot = join(__dirname, 'frontend', 'tests', 'features');

// Default BDD runs are skip-auth (see scripts/run-bdd.cjs). Real-auth features
// tagged with @auth-real must only execute when SKIP_AUTH=false.
const isSkipAuth = process.env.SKIP_AUTH !== 'false';

export default {
  paths: [join(featureRoot, '**', '*.feature')],
  require: [
    join(featureRoot, 'support', 'world.ts'),
    join(featureRoot, 'support', 'hooks.ts'),
    join(featureRoot, 'step_definitions', '**', '*.ts')
  ],
  import: ['tsx'],
  format: ['progress'],
  publishQuiet: true,
  tags: isSkipAuth ? 'not @auth-real' : ''
};
