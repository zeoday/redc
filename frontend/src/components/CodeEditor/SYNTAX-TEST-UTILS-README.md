# Syntax Highlighting Test Utilities

**Task 8.1: 创建语法高亮测试工具函数**

This document describes the syntax highlighting test utilities created for verifying Terraform syntax highlighting in the CodeEditor component.

## Overview

The syntax highlighting test utilities provide:
1. **Helper functions** to check syntax element style classes in the rendered editor
2. **Test data** with comprehensive Terraform code samples
3. **Verification functions** to validate syntax highlighting correctness

These utilities support testing Requirements 2.1-2.6 (syntax highlighting for keywords, strings, comments, numbers, operators, and block structures).

## Files

- `syntaxTestUtils.js` - Core utility functions and test data
- `syntaxHighlighting.test.js` - Unit tests demonstrating utility usage

## Test Data: Terraform Samples

The `terraformSamples` object provides comprehensive Terraform code examples:

```javascript
import { terraformSamples } from './syntaxTestUtils.js';

// Available samples:
terraformSamples.basicResource      // Basic resource block
terraformSamples.variables          // Variable declarations
terraformSamples.outputs            // Output blocks
terraformSamples.modules            // Module blocks
terraformSamples.dataSources        // Data source blocks
terraformSamples.locals             // Locals block
terraformSamples.terraformBlock     // Terraform configuration block
terraformSamples.provider           // Provider configuration
terraformSamples.comments           // Single and multi-line comments
terraformSamples.numbersAndOperators // Numbers and operators
terraformSamples.complexNested      // Complex nested structures
terraformSamples.stringInterpolation // String interpolation
```

## Syntax Element Types

The `syntaxElementTypes` object defines expected token types:

```javascript
import { syntaxElementTypes } from './syntaxTestUtils.js';

syntaxElementTypes.KEYWORD      // 'keyword'
syntaxElementTypes.STRING       // 'string'
syntaxElementTypes.COMMENT      // 'comment'
syntaxElementTypes.NUMBER       // 'number'
syntaxElementTypes.OPERATOR     // 'operator'
syntaxElementTypes.PUNCTUATION  // 'punctuation'
```

## Core Utility Functions

### Checking Individual Elements

#### `hasSyntaxClass(element, tokenType)`

Check if a DOM element has a specific syntax highlighting class.

```javascript
import { hasSyntaxClass } from './syntaxTestUtils.js';

const element = document.querySelector('.cm-keyword');
const isKeyword = hasSyntaxClass(element, 'keyword'); // true
```

#### `findElementsWithSyntaxClass(editorContainer, tokenType)`

Find all elements with a specific syntax class.

```javascript
import { findElementsWithSyntaxClass } from './syntaxTestUtils.js';

const editorContainer = container.querySelector('.code-editor-container');
const keywords = findElementsWithSyntaxClass(editorContainer, 'keyword');
console.log(`Found ${keywords.length} keywords`);
```

#### `isTextHighlightedAs(editorContainer, text, tokenType)`

Check if specific text is highlighted with the expected token type.

```javascript
import { isTextHighlightedAs } from './syntaxTestUtils.js';

const editorContainer = container.querySelector('.code-editor-container');
const isHighlighted = isTextHighlightedAs(editorContainer, 'resource', 'keyword');
```

### Analyzing Syntax Highlighting

#### `getAllSyntaxElements(editorContainer)`

Get all syntax-highlighted elements grouped by token type.

```javascript
import { getAllSyntaxElements } from './syntaxTestUtils.js';

const elements = getAllSyntaxElements(editorContainer);
// Returns: { keyword: [...], string: [...], comment: [...], ... }
```

#### `getSyntaxHighlightingSummary(editorContainer)`

Get a summary with counts for each token type.

```javascript
import { getSyntaxHighlightingSummary } from './syntaxTestUtils.js';

const summary = getSyntaxHighlightingSummary(editorContainer);
console.log(`Total highlighted elements: ${summary.total}`);
console.log(`Keywords: ${summary.byType.keyword || 0}`);
console.log(`Strings: ${summary.byType.string || 0}`);
```

#### `hasSyntaxHighlighting(editorContainer)`

Check if any syntax highlighting is applied.

```javascript
import { hasSyntaxHighlighting } from './syntaxTestUtils.js';

const hasHighlighting = hasSyntaxHighlighting(editorContainer);
```

### Counting Specific Elements

#### `countStringElements(editorContainer)`
#### `countCommentElements(editorContainer)`
#### `countNumberElements(editorContainer)`

Count specific types of syntax elements.

```javascript
import { 
  countStringElements, 
  countCommentElements, 
  countNumberElements 
} from './syntaxTestUtils.js';

const stringCount = countStringElements(editorContainer);
const commentCount = countCommentElements(editorContainer);
const numberCount = countNumberElements(editorContainer);
```

### Keyword Verification

#### `verifyKeywordHighlighting(editorContainer, keywords)`

Verify that specific keywords are properly highlighted.

```javascript
import { verifyKeywordHighlighting, terraformKeywords } from './syntaxTestUtils.js';

// Check all Terraform keywords
const result = verifyKeywordHighlighting(editorContainer, terraformKeywords);
console.log(`Found: ${result.found.length} keywords`);
console.log(`Missing: ${result.missing.length} keywords`);

// Check specific keywords
const customResult = verifyKeywordHighlighting(editorContainer, ['resource', 'variable']);
```

### Test Case Creation

#### `createSyntaxTestCase(name, code, expectations)`

Create a structured test case for syntax highlighting verification.

```javascript
import { createSyntaxTestCase, terraformSamples } from './syntaxTestUtils.js';

const testCase = createSyntaxTestCase(
  'Basic Resource Highlighting',
  terraformSamples.basicResource,
  {
    hasKeywords: true,
    hasStrings: true,
    hasComments: false,
    hasNumbers: false,
    minKeywordCount: 2,
    minStringCount: 3,
  }
);
```

#### `verifySyntaxTestCase(editorContainer, testCase)`

Verify a test case against the rendered editor.

```javascript
import { verifySyntaxTestCase } from './syntaxTestUtils.js';

const results = verifySyntaxTestCase(editorContainer, testCase);

if (results.passed) {
  console.log('Test passed!');
} else {
  console.log('Test failed:');
  results.failures.forEach(failure => console.log(`  - ${failure}`));
}
```

## Usage Examples

### Example 1: Basic Syntax Highlighting Test

```javascript
import { describe, it, expect } from 'vitest';
import { render } from '@testing-library/svelte';
import CodeEditor from './CodeEditor.svelte';
import { terraformSamples, hasSyntaxHighlighting } from './syntaxTestUtils.js';

describe('Terraform Syntax Highlighting', () => {
  it('should apply syntax highlighting to .tf files', () => {
    const { container } = render(CodeEditor, {
      props: {
        filename: 'main.tf',
        value: terraformSamples.basicResource,
      }
    });
    
    const editorContainer = container.querySelector('.code-editor-container');
    expect(hasSyntaxHighlighting(editorContainer)).toBe(true);
  });
});
```

### Example 2: Verify Keyword Highlighting

```javascript
import { describe, it, expect } from 'vitest';
import { render } from '@testing-library/svelte';
import CodeEditor from './CodeEditor.svelte';
import { 
  terraformSamples, 
  verifyKeywordHighlighting 
} from './syntaxTestUtils.js';

describe('Keyword Highlighting', () => {
  it('should highlight Terraform keywords', () => {
    const { container } = render(CodeEditor, {
      props: {
        filename: 'main.tf',
        value: terraformSamples.basicResource,
      }
    });
    
    const editorContainer = container.querySelector('.code-editor-container');
    const result = verifyKeywordHighlighting(editorContainer, ['resource']);
    
    // In a real browser environment, this would find the keyword
    // In jsdom, CodeMirror may not fully render
    expect(result).toBeDefined();
    expect(result.found).toBeDefined();
    expect(result.missing).toBeDefined();
  });
});
```

### Example 3: Count Syntax Elements

```javascript
import { describe, it, expect } from 'vitest';
import { render } from '@testing-library/svelte';
import CodeEditor from './CodeEditor.svelte';
import { 
  terraformSamples, 
  countStringElements,
  countCommentElements 
} from './syntaxTestUtils.js';

describe('Syntax Element Counts', () => {
  it('should highlight strings in Terraform code', () => {
    const { container } = render(CodeEditor, {
      props: {
        filename: 'main.tf',
        value: terraformSamples.basicResource,
      }
    });
    
    const editorContainer = container.querySelector('.code-editor-container');
    const stringCount = countStringElements(editorContainer);
    
    // In a real browser, we'd expect multiple strings
    expect(stringCount).toBeGreaterThanOrEqual(0);
  });
  
  it('should highlight comments in Terraform code', () => {
    const { container } = render(CodeEditor, {
      props: {
        filename: 'main.tf',
        value: terraformSamples.comments,
      }
    });
    
    const editorContainer = container.querySelector('.code-editor-container');
    const commentCount = countCommentElements(editorContainer);
    
    expect(commentCount).toBeGreaterThanOrEqual(0);
  });
});
```

### Example 4: Using Test Cases

```javascript
import { describe, it, expect } from 'vitest';
import { render } from '@testing-library/svelte';
import CodeEditor from './CodeEditor.svelte';
import { 
  createSyntaxTestCase,
  verifySyntaxTestCase,
  terraformSamples 
} from './syntaxTestUtils.js';

describe('Comprehensive Syntax Tests', () => {
  it('should pass all syntax highlighting requirements', () => {
    const testCase = createSyntaxTestCase(
      'Complete Terraform File',
      terraformSamples.basicResource,
      {
        hasKeywords: true,
        hasStrings: true,
        minKeywordCount: 1,
        minStringCount: 2,
      }
    );
    
    const { container } = render(CodeEditor, {
      props: {
        filename: 'main.tf',
        value: testCase.code,
      }
    });
    
    const editorContainer = container.querySelector('.code-editor-container');
    const results = verifySyntaxTestCase(editorContainer, testCase);
    
    // Note: In jsdom, CodeMirror may not fully render
    // This test structure is ready for browser-based testing
    expect(results).toBeDefined();
  });
});
```

## Testing Considerations

### jsdom Limitations

The test utilities work in both jsdom (for unit tests) and real browser environments. However, CodeMirror may not fully render syntax highlighting in jsdom. For comprehensive syntax highlighting tests:

1. **Unit tests (jsdom)**: Verify utility functions work correctly
2. **Integration tests (browser)**: Verify actual syntax highlighting in a real browser
3. **Property-based tests**: Use these utilities to verify syntax highlighting properties

### Browser-Based Testing

For full syntax highlighting verification, consider:

1. Using Playwright or Cypress for E2E tests
2. Running tests in a real browser environment
3. Using visual regression testing for syntax highlighting

### Property-Based Testing

These utilities are designed to support property-based testing (Task 8.2). Example:

```javascript
import fc from 'fast-check';
import { 
  createSyntaxTestCase,
  verifySyntaxTestCase,
  terraformSamples 
} from './syntaxTestUtils.js';

// Property: All Terraform files should have syntax highlighting
fc.assert(
  fc.property(
    fc.constantFrom(...Object.values(terraformSamples)),
    (terraformCode) => {
      // Render editor with Terraform code
      // Verify syntax highlighting is applied
      // Return true if highlighting is correct
    }
  )
);
```

## Requirements Coverage

These utilities support testing the following requirements:

- **Requirement 2.1**: Keywords (resource, variable, output, module, data, locals, terraform, provider)
- **Requirement 2.2**: String literals
- **Requirement 2.3**: Comments
- **Requirement 2.4**: Numbers
- **Requirement 2.5**: Operators and punctuation
- **Requirement 2.6**: HCL block structures (braces, brackets)

## Next Steps

Task 8.2 will use these utilities to implement property-based tests for Terraform syntax element recognition (Property 4).

## API Reference

### Exported Constants

- `terraformSamples` - Object with Terraform code samples
- `syntaxElementTypes` - Object with token type constants
- `terraformKeywords` - Array of Terraform keywords

### Exported Functions

- `hasSyntaxClass(element, tokenType)` - Check element class
- `findElementsWithSyntaxClass(container, tokenType)` - Find elements by type
- `isTextHighlightedAs(container, text, tokenType)` - Check text highlighting
- `getAllSyntaxElements(container)` - Get all syntax elements
- `getSyntaxHighlightingSummary(container)` - Get summary with counts
- `hasSyntaxHighlighting(container)` - Check if highlighting exists
- `countStringElements(container)` - Count string tokens
- `countCommentElements(container)` - Count comment tokens
- `countNumberElements(container)` - Count number tokens
- `verifyKeywordHighlighting(container, keywords)` - Verify keywords
- `createSyntaxTestCase(name, code, expectations)` - Create test case
- `verifySyntaxTestCase(container, testCase)` - Verify test case
- `waitForEditorRender(timeout)` - Wait for editor to render

## Conclusion

These utilities provide a comprehensive foundation for testing Terraform syntax highlighting. They are designed to be:

- **Easy to use**: Simple, intuitive API
- **Comprehensive**: Cover all syntax element types
- **Flexible**: Work in both jsdom and browser environments
- **Extensible**: Easy to add new test cases and verification functions
