/**
 * Adapter for FileSystem implementation
 * Provides a way to access the concrete implementation and ensures we can mock it in tests
 */
import { FileSystem } from './interfaces';
import { ConcreteFileSystem } from './FileSystem';
import { ReadFileOptions } from '../utils/fileReaderTypes';
import { Stats } from 'fs';

// This class is a simple adapter that ensures we can easily mock the FileSystem in tests
export class FileSystemAdapter implements FileSystem {
  private fs: FileSystem;

  constructor(customFs?: FileSystem) {
    this.fs = customFs || new ConcreteFileSystem();
  }

  // Forward all methods to the underlying implementation
  readFileContent = (filePath: string, options?: ReadFileOptions): Promise<string> =>
    this.fs.readFileContent(filePath, options);

  writeFile = (filePath: string, content: string): Promise<void> =>
    this.fs.writeFile(filePath, content);

  fileExists = (path: string): Promise<boolean> => this.fs.fileExists(path);

  mkdir = (dirPath: string, options?: { recursive?: boolean }): Promise<void> =>
    this.fs.mkdir(dirPath, options);

  readdir = (dirPath: string): Promise<string[]> => this.fs.readdir(dirPath);

  stat = (path: string): Promise<Stats> => this.fs.stat(path);

  access = (path: string, mode?: number): Promise<void> => this.fs.access(path, mode);

  getConfigDir = (): Promise<string> => this.fs.getConfigDir();

  getConfigFilePath = (): Promise<string> => this.fs.getConfigFilePath();
}
