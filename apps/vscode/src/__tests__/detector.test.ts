import { expect, test } from 'bun:test';
import { detectMachine, detectOS, detectTimezone } from '../detector';

// We only test pure, non-vscode-dependent logic here.
// Functions like detectOS and detectMachine use Node's
// built-in os module, so they're safe to test directly.

test('detectOS returns a known platform string', () => {
  const result = detectOS();
  expect(['linux', 'win32', 'darwin']).toContain(result);
});

test('detectMachine returns a non-empty string', () => {
  const result = detectMachine();
  expect(result.length).toBeGreaterThan(0);
});

test('detectTimezone returns a valid IANA timezone string', () => {
  const result = detectTimezone();
  expect(result).toMatch(/^[A-Za-z_]+(?:\/[A-Za-z_]+)*$/);
});
