# CodeEditor Integration Guide

## Overview

This guide explains how the CodeEditor component is integrated into the redc-gui application, specifically within the LocalTemplates component for editing Terraform template files.

## Integration Architecture

```
LocalTemplates.svelte
├── Template List (left panel)
│   └── Edit button → Opens editor modal
│
└── Template Editor Modal
    ├── File List (left sidebar)
    │   └── File selection buttons
    │
    └── CodeEditor Component (main area)
        ├── Detects file type (.tf, .tfvars)
        ├── Applies syntax highlighting
        └── Emits change events
```

## Data Flow

### 1. Opening the Editor

```javascript
// User clicks "Edit Template" button
async function openTemplateEditor(tmpl) {
  // Initialize editor state
  templateEditor = { 
    show: true, 
    name: tmpl.name, 
    files: {}, 
    active: '', 
    saving: false, 
    error: '' 
  };
  
  // Load template files from backend
  const files = await GetTemplateFiles(tmpl.name);
  
  // Update state with loaded files
  templateEditor = {
    ...templateEditor,
    files: files || {},
    active: Object.keys(files)[0] || '', // Select first file
  };
}
```

### 2. File Switching

```javascript
// User clicks a file in the file list
<button
  on:click={() => templateEditor = { ...templateEditor, active: fname }}
>
  {fname}
</button>

// CodeEditor receives new filename and value props
<CodeEditor
  filename={templateEditor.active}
  value={templateEditor.files[templateEditor.active]}
  on:change={handleChange}
/>
```

### 3. Content Changes

```javascript
// CodeEditor emits change event when user types
<CodeEditor
  filename={templateEditor.active}
  value={templateEditor.files[templateEditor.active]}
  on:change={(e) => {
    // Update the file content in state
    templateEditor.files[templateEditor.active] = e.detail;
    // Trigger reactivity
    templateEditor = templateEditor;
  }}
/>
```

### 4. Saving Changes

```javascript
// User clicks "Save" button
async function saveTemplateEditor() {
  templateEditor = { ...templateEditor, saving: true, error: '' };
  
  try {
    // Call backend API with all file contents
    await SaveTemplateFiles(templateEditor.name, templateEditor.files);
    templateEditor = { ...templateEditor, saving: false };
  } catch (e) {
    // Display error without closing modal
    templateEditor = { 
      ...templateEditor, 
      saving: false, 
      error: e.message || String(e) 
    };
  }
}
```

## State Management

### Template Editor State

```javascript
let templateEditor = {
  show: false,        // Whether modal is visible
  name: '',          // Template name
  files: {},         // Object mapping filename → content
  active: '',        // Currently selected filename
  saving: false,     // Whether save is in progress
  error: ''          // Error message (if any)
};
```

### Key Behaviors

1. **Unsaved Changes Preservation**: 
   - Changes are stored in `templateEditor.files` object
   - Switching files preserves unsaved changes in memory
   - Changes persist until modal is closed or saved

2. **File Selection**:
   - `templateEditor.active` tracks the current file
   - CodeEditor receives both `filename` and `value` props
   - Changing `active` triggers CodeEditor to update

3. **Error Handling**:
   - Errors are stored in `templateEditor.error`
   - Modal stays open on errors
   - User can retry save operation

## UI Layout

### Modal Structure

```html
<div class="modal">
  <!-- Header -->
  <div class="header">
    <h3>Edit Template</h3>
    <p>{templateEditor.name}</p>
    <button on:click={closeTemplateEditor}>Close</button>
    <button on:click={saveTemplateEditor}>Save</button>
  </div>
  
  <!-- Content -->
  <div class="content flex">
    <!-- File List Sidebar -->
    <div class="sidebar w-52">
      {#each Object.keys(templateEditor.files) as fname}
        <button 
          class:active={templateEditor.active === fname}
          on:click={() => selectFile(fname)}
        >
          {fname}
        </button>
      {/each}
    </div>
    
    <!-- Editor Area -->
    <div class="editor-area flex-1">
      {#if templateEditor.error}
        <div class="error">{templateEditor.error}</div>
      {/if}
      
      <CodeEditor
        filename={templateEditor.active}
        value={templateEditor.files[templateEditor.active]}
        on:change={handleChange}
      />
    </div>
  </div>
</div>
```

### Styling

The integration uses Tailwind CSS classes consistent with the application design:

- **Modal**: `bg-white rounded-xl shadow-xl max-w-4xl`
- **Sidebar**: `w-52 border-r border-gray-100`
- **Active file**: `bg-gray-900 text-white`
- **Inactive file**: `text-gray-600 hover:bg-gray-50`
- **Editor area**: `flex-1 p-4`

## API Integration

### Backend APIs Used

1. **GetTemplateFiles(templateName)**
   - Loads all files for a template
   - Returns: `{ [filename: string]: string }`
   - Called when opening editor

2. **SaveTemplateFiles(templateName, files)**
   - Saves all template files
   - Parameters: template name and files object
   - Called when user clicks Save

### Error Handling

```javascript
try {
  await SaveTemplateFiles(templateEditor.name, templateEditor.files);
  // Success - could show success message
} catch (e) {
  // Error - display in modal, don't close
  templateEditor = { 
    ...templateEditor, 
    error: e.message || String(e) 
  };
}
```

## Migration from Textarea

### Before (Original Implementation)

```svelte
<textarea
  bind:value={templateEditor.files[templateEditor.active]}
  class="w-full h-full border border-gray-300 rounded-md p-2"
/>
```

### After (CodeEditor Integration)

```svelte
<CodeEditor
  filename={templateEditor.active}
  value={templateEditor.files[templateEditor.active]}
  on:change={(e) => {
    templateEditor.files[templateEditor.active] = e.detail;
    templateEditor = templateEditor;
  }}
/>
```

### Key Differences

1. **Two-way binding → Event-based**:
   - Textarea: `bind:value` (automatic two-way binding)
   - CodeEditor: `on:change` event (explicit update)

2. **Reactivity trigger**:
   - Textarea: Automatic reactivity
   - CodeEditor: Must reassign object to trigger reactivity

3. **Syntax highlighting**:
   - Textarea: None
   - CodeEditor: Automatic for .tf/.tfvars files

## Best Practices

### 1. Always Trigger Reactivity

```javascript
// ❌ Wrong - doesn't trigger reactivity
templateEditor.files[templateEditor.active] = newContent;

// ✅ Correct - triggers reactivity
templateEditor.files[templateEditor.active] = newContent;
templateEditor = templateEditor;
```

### 2. Handle Errors Gracefully

```javascript
// ✅ Good - show error, keep modal open
catch (e) {
  templateEditor = { 
    ...templateEditor, 
    saving: false, 
    error: e.message 
  };
}

// ❌ Bad - close modal, lose changes
catch (e) {
  closeTemplateEditor();
  alert(e.message);
}
```

### 3. Preserve Unsaved Changes

```javascript
// ✅ Good - changes stay in memory
function selectFile(filename) {
  templateEditor = { ...templateEditor, active: filename };
}

// ❌ Bad - would lose unsaved changes
function selectFile(filename) {
  templateEditor = { 
    ...templateEditor, 
    active: filename,
    files: await GetTemplateFiles(templateEditor.name) // Reloads!
  };
}
```

### 4. Provide User Feedback

```javascript
// ✅ Good - show saving state
<button 
  on:click={saveTemplateEditor}
  disabled={templateEditor.saving}
>
  {templateEditor.saving ? 'Saving...' : 'Save'}
</button>

// ❌ Bad - no feedback
<button on:click={saveTemplateEditor}>
  Save
</button>
```

## Testing Integration

### Unit Tests

Test the integration points:

```javascript
test('updates file content on change event', () => {
  const { component } = render(LocalTemplates);
  
  // Open editor
  component.openTemplateEditor({ name: 'test' });
  
  // Simulate change event
  const editor = screen.getByRole('textbox');
  fireEvent.change(editor, { detail: 'new content' });
  
  // Verify state updated
  expect(component.templateEditor.files['main.tf']).toBe('new content');
});
```

### Integration Tests

Test the complete workflow:

```javascript
test('complete edit and save workflow', async () => {
  const { component } = render(LocalTemplates);
  
  // Open editor
  await component.openTemplateEditor({ name: 'test' });
  
  // Edit content
  const editor = screen.getByRole('textbox');
  fireEvent.change(editor, { detail: 'new content' });
  
  // Save
  const saveButton = screen.getByText('Save');
  await fireEvent.click(saveButton);
  
  // Verify API called
  expect(SaveTemplateFiles).toHaveBeenCalledWith(
    'test',
    { 'main.tf': 'new content' }
  );
});
```

## Troubleshooting

### Issue: Changes not persisting when switching files

**Cause**: Not triggering Svelte reactivity after updating nested object

**Solution**: Reassign the object after mutation
```javascript
templateEditor.files[filename] = newContent;
templateEditor = templateEditor; // Trigger reactivity
```

### Issue: Editor shows old content after file switch

**Cause**: CodeEditor not receiving updated props

**Solution**: Ensure both `filename` and `value` props are reactive
```svelte
<CodeEditor
  filename={templateEditor.active}
  value={templateEditor.files[templateEditor.active]}
  on:change={handleChange}
/>
```

### Issue: Save button stays disabled

**Cause**: `saving` state not reset after error

**Solution**: Always reset `saving` in finally block or catch
```javascript
try {
  await SaveTemplateFiles(...);
} catch (e) {
  // Handle error
} finally {
  templateEditor = { ...templateEditor, saving: false };
}
```

## Performance Considerations

### 1. Lazy Loading

CodeEditor modules are loaded on-demand:
- First editor open: ~200ms load time
- Subsequent opens: Instant (modules cached)

### 2. Large Files

CodeMirror 6 handles large files efficiently:
- Virtual scrolling for files > 10,000 lines
- Incremental parsing for syntax highlighting
- Efficient diff-based updates

### 3. Multiple Files

Switching between files is fast:
- No re-initialization needed
- Only content and syntax mode updated
- Changes preserved in memory

## Future Enhancements

Potential improvements to consider:

1. **Auto-save**: Automatically save changes after inactivity
2. **Undo/Redo**: Expose CodeMirror's undo/redo functionality
3. **Search/Replace**: Add search and replace functionality
4. **Code Folding**: Enable code folding for large files
5. **Validation**: Add Terraform syntax validation
6. **Formatting**: Add code formatting (terraform fmt)
7. **Autocomplete**: Add Terraform resource/variable autocomplete

## Related Documentation

- [CodeEditor README](./README.md) - Component documentation
- [Theme Customization](./theme.js) - Theme configuration
- [Syntax Test Utils](./SYNTAX-TEST-UTILS-README.md) - Testing utilities
- [Requirements](../../.kiro/specs/terraform-syntax-highlighting/requirements.md) - Feature requirements
- [Design](../../.kiro/specs/terraform-syntax-highlighting/design.md) - Design decisions
