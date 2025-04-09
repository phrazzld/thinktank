/**
 * Name generator module for thinktank
 *
 * Provides functions to generate user-friendly run names using a deterministic approach with
 * predefined word lists and fallback mechanisms when needed.
 */
import { logger } from './logger';

/**
 * List of adjectives used for generating friendly names
 */
export const ADJECTIVES: string[] = [
  'adaptable',
  'adventurous',
  'agile',
  'alert',
  'ambitious',
  'analytical',
  'artistic',
  'assertive',
  'attentive',
  'balanced',
  'bold',
  'brave',
  'bright',
  'brilliant',
  'calm',
  'capable',
  'careful',
  'charming',
  'cheerful',
  'clever',
  'collaborative',
  'committed',
  'compassionate',
  'confident',
  'considerate',
  'consistent',
  'courageous',
  'creative',
  'curious',
  'daring',
  'decisive',
  'dedicated',
  'deliberate',
  'determined',
  'diligent',
  'diplomatic',
  'disciplined',
  'dynamic',
  'eager',
  'efficient',
  'eloquent',
  'empathetic',
  'energetic',
  'enthusiastic',
  'excellent',
  'exceptional',
  'experienced',
  'expert',
  'focused',
  'friendly',
  'generous',
  'gentle',
  'genuine',
  'graceful',
  'happy',
  'harmonious',
  'helpful',
  'honest',
  'humble',
  'imaginative',
];

/**
 * List of nouns used for generating friendly names
 */
export const NOUNS: string[] = [
  'acorn',
  'anchor',
  'apple',
  'arrow',
  'aurora',
  'autumn',
  'badger',
  'bamboo',
  'bastion',
  'beacon',
  'blossom',
  'breeze',
  'brook',
  'butterfly',
  'canyon',
  'cardinal',
  'cascade',
  'cedar',
  'cheetah',
  'cherry',
  'citadel',
  'comet',
  'compass',
  'coral',
  'crystal',
  'cypress',
  'diamond',
  'dolphin',
  'eagle',
  'ember',
  'falcon',
  'feather',
  'firefly',
  'forest',
  'fountain',
  'galaxy',
  'garden',
  'gazelle',
  'geyser',
  'glacier',
  'griffin',
  'harbor',
  'heron',
  'horizon',
  'island',
  'jasmine',
  'lagoon',
  'lantern',
  'lotus',
  'marble',
  'meadow',
  'mercury',
  'nebula',
  'ocean',
  'orchid',
  'oasis',
  'panther',
  'phoenix',
  'quasar',
  'river',
];

// Removed Schema definition as we're using simple prompt-based approach

/**
 * Generates a fun, deterministic random name for a run
 * Format is 'adjective-noun' (e.g., clever-otter, swift-breeze)
 *
 * @returns A fun name string in the format 'adjective-noun'
 */
export function generateFunName(): string {
  // Basic check to prevent errors if lists are empty (should not happen in practice)
  if (ADJECTIVES.length === 0 || NOUNS.length === 0) {
    logger.debug('Adjective or Noun list is empty. Using fallback name.');
    // If lists are empty, use the fallback name generator
    return generateFallbackName();
  }

  // Select a random adjective and noun
  const adjIndex = Math.floor(Math.random() * ADJECTIVES.length);
  const nounIndex = Math.floor(Math.random() * NOUNS.length);

  const adjective = ADJECTIVES[adjIndex];
  const noun = NOUNS[nounIndex];

  // Combine them with a hyphen
  return `${adjective}-${noun}`;
}

/**
 * Generates a fallback name based on timestamp
 * Used when the Gemini API call fails
 *
 * @returns A timestamp-based name (e.g., "run-20250101-123045")
 */
export function generateFallbackName(): string {
  const now = new Date();
  const timestamp = now
    .toISOString()
    .replace(/[-:T.Z]/g, '') // Remove separators
    .substring(0, 14); // YYYYMMDDHHmmss
  return `run-${timestamp.substring(0, 8)}-${timestamp.substring(8)}`; // Format as run-YYYYMMDD-HHmmss
}
