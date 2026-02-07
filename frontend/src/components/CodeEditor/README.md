# CodeEditor Component

## Overview

The `CodeEditor` component is a Svelte-based code editor with Terraform syntax highlighting, built on top of CodeMirror 6. It provides a professional code editing experience with features like line numbers, syntax highlighting, indentation support, and a custom theme that matches the application's Tailwind CSS design system.

## Features

- **Terraform Syntax Highlighting**: Automatically detects `.tf` and `.tfvars` files and applies Terraform syntax highlighting
- **Plain Text Mode**: Falls back to plain text mode for non-Terraform files
- **Line Numbers**: Displays line numbers for all lines
- **Custom Theme**: Uses a custom theme that matches the application's Tailwind CSS design system
- **Lazy Loading**: CodeMirror modules are loaded on-demand to improve initial page load performance
- **Error Handling**: Gracefully falls back to a textarea if CodeMirror fails to load
- **Responsive**: Adapts to container size and provides proper scrolling
- **Read-only Mode**: Supports read-only mode for viewing code without editing

## Installation

The component requires the following dependencies:

```bash
npm install codemirror @codemirror/state @codemirror/language @codemirror/view @codemirror/commands codemirror-lang-terraform
```

## Usage

### Basic Usage

```svelte
<script>
  import CodeEditor from './components/CodeEditor/CodeEditor.svelte';
  
  let filename = 'main.tf';
  let content = 'resource "aws_instance" "example" {\n  ami = "ami-12345"\n}';
  
  function handleChange(event) {
    content = event.detail;
    console.log('Content changed:', content);
  }
</script>

<CodeEditor 
  {filename} 
  value={content} 
  on:change={handleChange} 
/>
```

### Read-only Mode

```svelte
<CodeEditor 
  filename="main.tf" 
  value={content} 
  readonly={true}
/>
```

### With File Switching

```svelte
<script>
  import CodeEditor from './components/CodeEditor/CodeEditor.svelte';
  
  let files = {
    'main.tf': 'resource "aws_instance" "example" {...}',
    'variables.tf': 'variable "region" {...}',
    'outputs.tf': 'output "instance_ip" {...}'
  };
  
  let activeFile = 'main.tf';
  
  function handleChange(event) {
    files[activeFile] = event.detail;
  }
</script>

<div class="file-list">
  {#each Object.keys(files) as filename}
    <button on:click={() => activeFile = filename}>
      {filename}
    </button>
  {/each}
</div>

<CodeEditor 
  filename={activeFile} 
  value={files[activeFile]} 
  on:change={handleChange} 
/>
```

## Props

| Prop | Type | Default | Description |
|------|------|---------|-------------|
| `filename` | `string` | `''` | The name of the file being edited. Used to determine syntax highlighting mode. |
| `value` | `string` | `''` | The content of the file. |
| `readonly` | `boolean` | `false` | Whether the editor is read-only. |

## Events

| Event | Detail Type | Description |
|-------|-------------|-------------|
| `change` | `string` | Fired when the editor content changes. The detail contains the new content. |

## File Type Detection

The component automatically detects Terraform files based on their extension:

- **Terraform files**: `.tf`, `.tfvars` → Terraform syntax highlighting
- **Other files**: Plain text mode (no syntax highlighting)

## Theme Customization

The editor uses a custom theme defined in `theme.js` that matches the application's Tailwind CSS design system. The theme includes:

- **Editor colors**: Gray-50 background, Gray-900 text
- **Gutter colors**: Gray-100 background, Gray-500 text
- **Selection colors**: Blue-100/Blue-200 backgrounds
- **Syntax highlighting**: Custom colors for keywords, strings, comments, numbers, operators, etc.

To customize the theme, edit the `theme.js` file.

## Performance

The component implements several performance optimizations:

1. **Lazy Loading**: CodeMirror modules are loaded on-demand using dynamic imports
2. **Loading Indicator**: Shows a loading spinner while modules are being loaded
3. **Efficient Updates**: Only updates the editor when props actually change
4. **Virtual Scrolling**: CodeMirror 6 includes built-in virtual scrolling for large files

## Error Handling

The component includes robust error handling:

1. **Module Loading Failure**: If CodeMirror modules fail to load, the component falls back to a plain textarea
2. **Initialization Failure**: If the editor fails to initialize, an error message is displayed and the fallback textarea is used
3. **Syntax Loading Failure**: If Terraform syntax highlighting fails to load, the editor continues to work in plain text mode

## Testing

The component includes comprehensive tests:

- **Unit Tests**: Test specific functionality like file type detection, content changes, etc.
- **Integration Tests**: Test the component's integration with parent components
- **Property-Based Tests**: Verify universal properties across all inputs
- **Syntax Highlighting Tests**: Verify that Terraform syntax elements are correctly highlighted

See the test files in the same directory for examples.

## Architecture

### Component Lifecycle

1. **Mount**: 
   - Initialize editor container
   - Load CodeMirror modules (lazy loading)
   - Create editor instance with appropriate extensions
   - Apply custom theme

2. **Update**:
   - When `filename` changes: Re-initialize editor with new syntax mode
   - When `value` changes: Update editor content if different from current content

3. **Destroy**:
   - Clean up editor instance
   - Release resources

### Extension Configuration

The editor uses the following CodeMirror extensions:

- `basicSetup`: Provides basic editor features (line numbers, bracket matching, etc.)
- `editorTheme`: Custom theme matching Tailwind CSS design
- `indentUnit`: Configures indentation (2 spaces)
- `keymap`: Configures Tab key behavior
- `EditorView.updateListener`: Listens for content changes
- `terraform()`: Terraform syntax highlighting (for .tf/.tfvars files)
- `EditorState.readOnly`: Read-only mode (when enabled)

## Integration with LocalTemplates

The CodeEditor is integrated into the `LocalTemplates` component to provide a professional editing experience for template files:

```svelte
<!-- In LocalTemplates.svelte -->
<CodeEditor
  filename={templateEditor.active}
  value={templateEditor.files[templateEditor.active]}
  on:change={(e) => {
    templateEditor.files[templateEditor.active] = e.detail;
    templateEditor = templateEditor;
  }}
/>
```

The integration maintains all existing functionality:
- File switching preserves unsaved changes
- Save button calls the existing `SaveTemplateFiles` API
- Error handling uses the existing error display mechanism

## Browser Compatibility

The component is compatible with modern browsers that support:
- ES6 modules
- Dynamic imports
- CSS Grid and Flexbox
- Modern JavaScript features (async/await, etc.)

Tested on:
- Chrome 90+
- Firefox 88+
- Safari 14+
- Edge 90+

## Troubleshooting

### Editor not loading

If the editor shows a loading spinner indefinitely:
1. Check the browser console for errors
2. Verify that CodeMirror dependencies are installed
3. Check network tab for failed module loads

### Syntax highlighting not working

If Terraform syntax highlighting is not applied:
1. Verify the filename has `.tf` or `.tfvars` extension
2. Check the browser console for syntax loading errors
3. Verify `codemirror-lang-terraform` is installed

### Fallback textarea appears

If the fallback textarea is shown instead of the editor:
1. Check the error message displayed above the textarea
2. Verify all dependencies are installed correctly
3. Check the browser console for detailed error information

## Requirements Mapping

This component satisfies the following requirements from the specification:

- **Requirement 1**: Terraform file identification (1.1, 1.2, 1.3)
- **Requirement 2**: Syntax highlighting (2.1-2.6)
- **Requirement 3**: Code editor functionality (3.1-3.5)
- **Requirement 4**: Maintain existing functionality (4.1-4.4)
- **Requirement 5**: UI consistency (5.1-5.4)
- **Requirement 6**: Performance requirements (6.1-6.3)
- **Requirement 7**: Error handling (7.1-7.3)
- **Requirement 8**: Code editor library integration (8.1-8.4)

## Contributing

When contributing to this component:

1. **Add tests**: All new features should include unit tests and/or property-based tests
2. **Update documentation**: Update this README and inline comments
3. **Follow style guide**: Use consistent formatting and naming conventions
4. **Test thoroughly**: Test with various file types and edge cases
5. **Check performance**: Ensure changes don't negatively impact performance

## License

This component is part of the redc-gui application and follows the same license.
