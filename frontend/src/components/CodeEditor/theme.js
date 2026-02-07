/**
 * Custom CodeMirror theme matching Tailwind CSS design system
 * 
 * This theme provides a consistent visual experience with the application's
 * Tailwind-based design, using colors from the gray palette and proper
 * typography settings.
 * 
 * Requirements: 5.1, 5.2, 5.3, 8.4
 */

import { EditorView } from 'codemirror';

/**
 * Custom theme configuration for CodeMirror editor
 * Matches the application's Tailwind CSS design system
 */
export const customTheme = EditorView.theme({
  // Editor container
  '&': {
    backgroundColor: '#f9fafb', // gray-50
    color: '#111827', // gray-900
    height: '100%',
  },

  // Editor content area
  '.cm-content': {
    fontFamily: 'ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace',
    fontSize: '12px',
    lineHeight: '1.5',
    caretColor: '#111827', // gray-900
    padding: '8px 0',
  },

  // Scroller
  '.cm-scroller': {
    fontFamily: 'ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace',
    lineHeight: '1.5',
  },

  // Line numbers gutter
  '.cm-gutters': {
    backgroundColor: '#f3f4f6', // gray-100
    color: '#6b7280', // gray-500
    border: 'none',
    borderRight: '1px solid #e5e7eb', // gray-200
    paddingRight: '8px',
    minWidth: '40px',
  },

  // Active line gutter (line number of current line)
  '.cm-activeLineGutter': {
    backgroundColor: '#e5e7eb', // gray-200
    color: '#374151', // gray-700
  },

  // Active line background
  '.cm-activeLine': {
    backgroundColor: '#f3f4f6', // gray-100
  },

  // Selection background (when not focused)
  '.cm-selectionBackground': {
    backgroundColor: '#dbeafe !important', // blue-100
  },

  // Selection background (when focused)
  '&.cm-focused .cm-selectionBackground': {
    backgroundColor: '#bfdbfe !important', // blue-200
  },

  // Cursor
  '.cm-cursor': {
    borderLeftColor: '#111827', // gray-900
    borderLeftWidth: '2px',
  },

  // Matching brackets
  '.cm-matchingBracket': {
    backgroundColor: '#fef3c7', // amber-100
    outline: '1px solid #fbbf24', // amber-400
  },

  // Non-matching brackets
  '.cm-nonmatchingBracket': {
    backgroundColor: '#fee2e2', // red-100
    outline: '1px solid #ef4444', // red-500
  },

  // Search match
  '.cm-searchMatch': {
    backgroundColor: '#fef3c7', // amber-100
    outline: '1px solid #f59e0b', // amber-500
  },

  // Selected search match
  '.cm-searchMatch.cm-searchMatch-selected': {
    backgroundColor: '#fcd34d', // amber-300
  },

  // Line wrapping
  '.cm-line': {
    padding: '0 4px',
  },

  // Placeholder text
  '.cm-placeholder': {
    color: '#9ca3af', // gray-400
    fontStyle: 'italic',
  },

  // Focused state
  '&.cm-focused': {
    outline: 'none',
  },

  // Panels (like search panel)
  '.cm-panels': {
    backgroundColor: '#f9fafb', // gray-50
    color: '#111827', // gray-900
    borderTop: '1px solid #e5e7eb', // gray-200
  },

  '.cm-panels-top': {
    borderBottom: '1px solid #e5e7eb', // gray-200
  },

  '.cm-panels-bottom': {
    borderTop: '1px solid #e5e7eb', // gray-200
  },

  // Tooltip
  '.cm-tooltip': {
    backgroundColor: '#1f2937', // gray-800
    color: '#f9fafb', // gray-50
    border: '1px solid #374151', // gray-700
    borderRadius: '0.375rem', // rounded-md
    padding: '4px 8px',
    fontSize: '12px',
  },

  '.cm-tooltip-autocomplete': {
    '& > ul': {
      fontFamily: 'ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace',
      fontSize: '12px',
    },
    '& > ul > li': {
      padding: '2px 8px',
    },
    '& > ul > li[aria-selected]': {
      backgroundColor: '#3b82f6', // blue-500
      color: '#ffffff',
    },
  },
});

/**
 * Syntax highlighting styles for Terraform and general code
 * These styles define colors for different token types
 */
export const syntaxHighlighting = EditorView.theme({
  // Keywords (resource, variable, output, etc.)
  '.cm-keyword': {
    color: '#7c3aed', // violet-600
    fontWeight: '600',
  },

  // Strings
  '.cm-string': {
    color: '#059669', // emerald-600
  },

  // Comments
  '.cm-comment': {
    color: '#6b7280', // gray-500
    fontStyle: 'italic',
  },

  // Numbers
  '.cm-number': {
    color: '#dc2626', // red-600
  },

  // Booleans
  '.cm-bool': {
    color: '#dc2626', // red-600
    fontWeight: '600',
  },

  // Operators
  '.cm-operator': {
    color: '#374151', // gray-700
  },

  // Punctuation (braces, brackets, parentheses)
  '.cm-punctuation': {
    color: '#4b5563', // gray-600
  },

  // Property names
  '.cm-propertyName': {
    color: '#0891b2', // cyan-600
  },

  // Variable names
  '.cm-variableName': {
    color: '#111827', // gray-900
  },

  // Type names
  '.cm-typeName': {
    color: '#ea580c', // orange-600
    fontWeight: '600',
  },

  // Function/method names
  '.cm-function': {
    color: '#2563eb', // blue-600
  },

  // Definition (when defining a variable, function, etc.)
  '.cm-definition': {
    color: '#111827', // gray-900
    fontWeight: '600',
  },

  // Invalid/error tokens
  '.cm-invalid': {
    color: '#dc2626', // red-600
    textDecoration: 'underline wavy #dc2626',
  },

  // Meta (special syntax elements)
  '.cm-meta': {
    color: '#7c3aed', // violet-600
  },

  // Tags (for markup languages)
  '.cm-tag': {
    color: '#059669', // emerald-600
  },

  // Attributes
  '.cm-attribute': {
    color: '#0891b2', // cyan-600
  },
});

/**
 * Combined theme with both editor styling and syntax highlighting
 */
export const editorTheme = [customTheme, syntaxHighlighting];
