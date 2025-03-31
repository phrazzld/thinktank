/**
 * Registry for LLM providers in the thinktank application
 * 
 * The registry is responsible for managing the registration and lookup of LLM providers.
 * It allows the application to dynamically register and use different LLM providers.
 */
import { LLMProvider } from '../atoms/types';

/**
 * Error thrown when there are issues with the provider registry
 */
export class ProviderRegistryError extends Error {
  constructor(message: string) {
    super(message);
    this.name = 'ProviderRegistryError';
  }
}

/**
 * The LLM provider registry class
 */
class LLMProviderRegistry {
  private providers: Map<string, LLMProvider>;
  
  constructor() {
    this.providers = new Map<string, LLMProvider>();
  }
  
  /**
   * Registers a provider with the registry
   * 
   * @param provider - The provider to register
   * @throws {ProviderRegistryError} If the provider is invalid or already registered
   */
  registerProvider(provider: LLMProvider): void {
    if (!provider) {
      throw new ProviderRegistryError('Cannot register undefined or null provider');
    }
    
    if (!provider.providerId) {
      throw new ProviderRegistryError('Provider must have a providerId');
    }
    
    if (this.providers.has(provider.providerId)) {
      throw new ProviderRegistryError(`Provider with ID '${provider.providerId}' is already registered`);
    }
    
    this.providers.set(provider.providerId, provider);
  }
  
  /**
   * Gets a provider by its ID
   * 
   * @param providerId - The ID of the provider to get
   * @returns The provider, or undefined if not found
   */
  getProvider(providerId: string): LLMProvider | undefined {
    return this.providers.get(providerId);
  }
  
  /**
   * Checks if a provider is registered
   * 
   * @param providerId - The ID of the provider to check
   * @returns True if the provider is registered, false otherwise
   */
  hasProvider(providerId: string): boolean {
    return this.providers.has(providerId);
  }
  
  /**
   * Gets all registered provider IDs
   * 
   * @returns Array of provider IDs
   */
  getProviderIds(): string[] {
    return Array.from(this.providers.keys());
  }
  
  /**
   * Gets all registered providers
   * 
   * @returns Array of providers
   */
  getAllProviders(): LLMProvider[] {
    return Array.from(this.providers.values());
  }
  
  /**
   * Clears all registered providers
   * Primarily used for testing
   */
  clear(): void {
    this.providers.clear();
  }
}

// Create a singleton instance of the registry
const registry = new LLMProviderRegistry();

/**
 * Registers a provider with the registry
 * 
 * @param provider - The provider to register
 * @throws {ProviderRegistryError} If the provider is invalid or already registered
 */
export function registerProvider(provider: LLMProvider): void {
  registry.registerProvider(provider);
}

/**
 * Gets a provider by its ID
 * 
 * @param providerId - The ID of the provider to get
 * @returns The provider, or undefined if not found
 */
export function getProvider(providerId: string): LLMProvider | undefined {
  return registry.getProvider(providerId);
}

/**
 * Checks if a provider is registered
 * 
 * @param providerId - The ID of the provider to check
 * @returns True if the provider is registered, false otherwise
 */
export function hasProvider(providerId: string): boolean {
  return registry.hasProvider(providerId);
}

/**
 * Gets all registered provider IDs
 * 
 * @returns Array of provider IDs
 */
export function getProviderIds(): string[] {
  return registry.getProviderIds();
}

/**
 * Gets all registered providers
 * 
 * @returns Array of providers
 */
export function getAllProviders(): LLMProvider[] {
  return registry.getAllProviders();
}

/**
 * Clears all registered providers
 * Primarily used for testing
 */
export function clearRegistry(): void {
  registry.clear();
}