import assert from 'node:assert/strict';
import test from 'node:test';

import {formatSize, keysOf, LANGUAGES, summary, translator} from './i18n.js';

test('every language has the same keys', () => {
    const [first, ...rest] = LANGUAGES;
    for (const lang of rest) {
        assert.deepEqual(keysOf(lang).sort(), keysOf(first).sort(), lang);
    }
});

test('unknown languages fall back to English', () => {
    assert.equal(translator('fr')('open'), 'Open');
    assert.equal(translator('es-ES')('open'), 'Abrir');
});

test('formatSize', () => {
    assert.equal(formatSize(512), '512 B');
    assert.equal(formatSize(1536), '1.5 KB');
    assert.equal(formatSize(5 * 1024 * 1024), '5.0 MB');
});

test('summary', () => {
    const t = translator('es');
    assert.equal(summary({has_text: true, file_count: 2}, t), 'Texto · 2 ficheros');
    assert.equal(summary({has_text: false, file_count: 1}, t), '1 fichero');
});
