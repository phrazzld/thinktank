/**
 * Unit tests for type definitions
 */
import { LLMProvider, LLMAvailableModel } from '../types';

describe('Type Definitions', () => {
  describe('LLMProvider Interface', () => {
    it('should allow implementing generate method', () => {
      // This test verifies that we can create a valid implementation of the interface
      const provider: LLMProvider = {
        providerId: 'test-provider',
        async generate(prompt, modelId) {
          return {
            provider: this.providerId,
            modelId,
            text: `Test response for ${prompt}`
          };
        }
      };

      // The test passes if TypeScript accepts this implementation
      expect(provider.providerId).toBe('test-provider');
    });

    it('should allow implementing both generate and listModels methods', () => {
      // This test verifies that we can implement the optional listModels method
      const providerWithListing: LLMProvider = {
        providerId: 'test-provider-with-listing',
        async generate(prompt, modelId) {
          return {
            provider: this.providerId,
            modelId,
            text: `Test response for ${prompt}`
          };
        },
        // Including the listModels method
        async listModels() {
          return [
            { id: 'model-1' },
            { id: 'model-2', description: 'Test Model 2' }
          ];
        }
      };

      // The test passes if TypeScript accepts this implementation
      expect(providerWithListing.providerId).toBe('test-provider-with-listing');
      expect(typeof providerWithListing.listModels).toBe('function');
    });
  });

  describe('LLMAvailableModel Interface', () => {
    it('should allow creating model information objects', () => {
      // This test verifies that we can create valid LLMAvailableModel objects
      const models: LLMAvailableModel[] = [
        { id: 'basic-model' },
        { id: 'detailed-model', description: 'A model with description' }
      ];

      expect(models).toHaveLength(2);
      expect(models[0].id).toBe('basic-model');
      expect(models[0].description).toBeUndefined();
      expect(models[1].id).toBe('detailed-model');
      expect(models[1].description).toBe('A model with description');
    });
  });
});