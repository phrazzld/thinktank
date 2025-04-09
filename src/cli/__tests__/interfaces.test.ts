import {
  CommandHandler,
  RunCommandHandler,
  ModelsListCommandHandler,
  ConfigCommandHandler
} from '../interfaces';
import {
  RunCommandOptions,
  ModelsListCommandOptions,
  ConfigCommandOptions
} from '../types';

describe('CLI Command Handler Interfaces', () => {
  // Mock implementations for testing interface compliance
  class MockRunCommandHandler implements RunCommandHandler {
    public executeCount = 0;
    public lastOptions: RunCommandOptions | null = null;
    
    async execute(options: RunCommandOptions): Promise<void> {
      this.executeCount++;
      this.lastOptions = options;
    }
  }

  class MockModelsListCommandHandler implements ModelsListCommandHandler {
    public executeCount = 0;
    public lastOptions: ModelsListCommandOptions | null = null;
    
    async execute(options: ModelsListCommandOptions): Promise<void> {
      this.executeCount++;
      this.lastOptions = options;
    }
  }

  class MockConfigCommandHandler implements ConfigCommandHandler {
    public executeCount = 0;
    public lastOptions: ConfigCommandOptions | null = null;
    
    async execute(options: ConfigCommandOptions): Promise<void> {
      this.executeCount++;
      this.lastOptions = options;
    }
  }

  describe('Interface Implementation', () => {
    it('should allow implementation of RunCommandHandler', async () => {
      const handler = new MockRunCommandHandler();
      const options: RunCommandOptions = {
        promptFile: 'prompt.md',
        contextPaths: ['./context'],
        groupName: 'test-group'
      };
      
      await handler.execute(options);
      
      expect(handler.executeCount).toBe(1);
      expect(handler.lastOptions).toEqual(options);
    });

    it('should allow implementation of ModelsListCommandHandler', async () => {
      const handler = new MockModelsListCommandHandler();
      const options: ModelsListCommandOptions = {
        provider: 'openai',
        detailed: true
      };
      
      await handler.execute(options);
      
      expect(handler.executeCount).toBe(1);
      expect(handler.lastOptions).toEqual(options);
    });

    it('should allow implementation of ConfigCommandHandler', async () => {
      const handler = new MockConfigCommandHandler();
      const options: ConfigCommandOptions = {
        action: 'set',
        key: 'defaultProvider',
        value: 'openai'
      };
      
      await handler.execute(options);
      
      expect(handler.executeCount).toBe(1);
      expect(handler.lastOptions).toEqual(options);
    });
  });

  describe('Type Safety', () => {
    it('should enforce required properties in RunCommandOptions', () => {
      // This test verifies at compile time that promptFile is required
      const options: RunCommandOptions = {
        promptFile: 'prompt.md'
      };
      
      expect(options.promptFile).toBeDefined();
      
      // TypeScript would catch this error at compile time:
      // const invalidOptions: RunCommandOptions = {}; // Error: Property 'promptFile' is missing
    });
    
    it('should work with the generic CommandHandler interface', async () => {
      // Create a handler with the generic interface
      const handler: CommandHandler<RunCommandOptions> = new MockRunCommandHandler();
      const options: RunCommandOptions = {
        promptFile: 'prompt.md'
      };
      
      await handler.execute(options);
      
      // TypeScript would catch this error at compile time:
      // await handler.execute({}); // Error: Property 'promptFile' is missing
    });
  });
});
