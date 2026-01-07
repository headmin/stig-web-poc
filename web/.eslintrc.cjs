module.exports = {
  root: true,
  env: {
    browser: true,
    es2021: true,
    node: true,
  },
  extends: [
    'eslint:recommended',
    'plugin:@typescript-eslint/recommended',
    'plugin:vue/vue3-essential',  // Less strict than vue3-recommended
  ],
  parser: 'vue-eslint-parser',
  parserOptions: {
    parser: '@typescript-eslint/parser',
    ecmaVersion: 'latest',
    sourceType: 'module',
  },
  plugins: ['@typescript-eslint', 'import'],
  rules: {
    // Critical: Catch import order issues (imports before code)
    'import/first': 'error',
    
    // Relaxed: Don't require newlines between import groups
    'import/order': 'off',
    
    // Relaxed: Allow use-before-define for functions (hoisted)
    '@typescript-eslint/no-use-before-define': ['error', {
      functions: false,  // Functions are hoisted
      classes: true,
      variables: true,
    }],
    
    // Relaxed: Allow any in type definitions
    '@typescript-eslint/no-explicit-any': 'off',
    '@typescript-eslint/ban-types': 'off',
    
    // Vue: Only essential rules, not style rules
    'vue/multi-word-component-names': 'off',
  },
}
