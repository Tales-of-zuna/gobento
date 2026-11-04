import { describe, expect, it } from 'vitest';
import { titleInSentence } from './kinds';

describe('titleInSentence', () => {
  it('lower-cases the first letter but keeps a leading acronym', () => {
    expect(titleInSentence('Database')).toBe('database');
    expect(titleInSentence('Data access')).toBe('data access');
    expect(titleInSentence('HTTP framework')).toBe('HTTP framework');
  });
});
