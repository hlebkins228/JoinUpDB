import { createHash, randomUUID } from 'node:crypto';
import { mkdirSync, readFileSync, readdirSync, writeFileSync, copyFileSync, existsSync, rmSync } from 'node:fs';
import { join } from 'node:path';

const output = process.argv[2];
const target = join(output, 'allure-results');
rmSync(target, { recursive: true, force: true });
mkdirSync(target, { recursive: true });
const names = {
  unit: [...readFileSync('internal/service/review_allure_test.go', 'utf8').matchAll(/r\.NewTest\("([^"]+)"/g)].map(match => match[1]),
  integration: ['Review creation and reading', 'Duplicate review is rejected', 'Future event review is rejected'],
  e2e: ['Register, login, create, read and delete event'],
};
function status(stage) {
  const path = join(output, `${stage}.status`);
  return existsSync(path) ? readFileSync(path, 'utf8').trim() : 'skipped';
}
function result(stage, name, state, message) {
  const uuid = randomUUID();
  const fullName = `${stage}/${name}`;
  const log = `${stage}.log`;
  const hasLog = existsSync(join(output, log));
  if (hasLog) copyFileSync(join(output, log), join(target, log));
  writeFileSync(join(target, `${uuid}-result.json`), JSON.stringify({
    uuid, name, fullName,
    historyId: createHash('md5').update(fullName).digest('hex'),
    status: state, statusDetails: { message }, start: Date.now(), stop: Date.now(),
    labels: [{ name: 'parentSuite', value: stage }, { name: 'suite', value: stage }],
    attachments: hasLog ? [{ name: 'Process log', source: log, type: 'text/plain' }] : [],
  }));
}
for (const [stage, expected] of Object.entries(names)) {
  const folder = join(output, 'results', stage);
  const actual = [];
  if (existsSync(folder)) {
    for (const name of readdirSync(folder)) {
      const source = join(folder, name);
      if (name.endsWith('-result.json')) {
        let test;
        try {
          test = JSON.parse(readFileSync(source, 'utf8'));
        } catch {
          result(stage, 'Incomplete test result', 'broken', `Could not read ${name} after an interrupted run`);
          continue;
        }
        test.labels.push({ name: 'parentSuite', value: stage });
        actual.push(test);
        writeFileSync(join(target, name), JSON.stringify(test));
      } else {
        copyFileSync(source, join(target, name));
      }
    }
  }
  for (const name of expected) {
    if (!actual.some(test => test.name === name)) {
      result(stage, name, 'skipped', `Stage ${stage}: ${status(stage)}; test was not executed`);
    }
  }
  if (status(stage) !== 'passed' && status(stage) !== 'skipped' && !actual.some(test => ['failed', 'broken'].includes(test.status))) {
    result(stage, `${stage} process failed`, 'failed', 'See process log for setup, build or interruption errors');
  }
}
for (const stage of ['cleanup', 'replay', 'capture']) {
  if (status(stage) === 'failed') result(stage, `${stage} failed`, 'failed', `See ${stage} log`);
}
writeFileSync(join(target, 'executor.json'), JSON.stringify({
  name: process.env.GITHUB_ACTIONS ? 'GitHub Actions' : 'Docker Compose', type: 'github',
  buildOrder: Number(process.env.BUILD_NUMBER || 0), buildName: `Run ${process.env.BUILD_NUMBER || 'local'}`,
  buildUrl: process.env.BUILD_URL || '', reportName: 'JoinUp: unit → integration → e2e',
}));
