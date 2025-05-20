// Local commitlint configuration that extends .commitlintrc.yml
// with baseline commit validation
module.exports = {
  extends: ['./.commitlintrc.yml'],
  // Only validate commits after the baseline commit when conventional commits were established
  ignores: [
    (commit) => {
      // This is a simplified ignores function for commit-msg hooks
      // It will be skipped during pre-commit hook validation
      // The full validation with proper baseline is done in CI
      return false;
    }
  ]
};
