# CodeEditor Quick Reference

## Basic Usage

```svelte
<script>
  import CodeEditor from './components/CodeEditor/CodeEditor.svelte';
  
  let filename = 'main.tf';
  let content = '';
</script>

<CodeEditor 
  {filename} 
  value={content} 
  on:change={(e) => content = e.detail} 
/>
```

## Props

| Prop | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `filename` | string | No | `''` | File name (determines syntax highlighting) |
| `value` | string | No | `''` | File content |
| `readonly` | boolean | No | `false` | Read-only mode |

## Events

| Event | Detail | Description |
|-------|--------|-------------|
| `change` | `string` | Fired when content changes |

## File Type Detection

- `.tf` → Terraform syntax highlighting
- `.tfvars` → Terraform syntax highlighting  
- Other → Plain text mode

## Common Patterns

### File Switching

```svelte
<script>
  let files = { 'main.tf': '...', 'vars.tf': '...' };
  let active = 'main.tf';
</script>

<CodeEditor 
  filename={active} 
  value={files[active]} 
  on:change={(e) => {
    files[active] = e.detail;
    files = files; // Trigger reactivity
  }}
/>
```

### Read-only Viewer

```svelte
<CodeEditor 
  filename="main.tf" 
  value={content} 
  readonly={true}
/>
```

### With Loading State

```svelte
<script>
  let loading = true;
  let content = '';
  
  onMount(async () => {
    content = await loadFile();
    loading = false;
  });
</script>

{#if loading}
  <div>Loading...</div>
{:else}
  <CodeEditor filename="main.tf" value={content} on:change={handleChange} />
{/if}
```

## Styling

The editor automatically matches your Tailwind CSS theme. To customize:

1. Edit `theme.js` for colors and styles
2. Use CSS classes on the wrapper div
3. Set height via parent container

```svelte
<div class="h-96 border rounded">
  <CodeEditor {filename} {value} on:change={handleChange} />
</div>
```

## Error Handling

The component handles errors gracefully:

- **Module load failure** → Falls back to textarea
- **Initialization failure** → Shows error + textarea
- **Syntax load failure** → Plain text mode

No special error handling needed in your code.

## Performance Tips

1. **Lazy loading**: Modules load on first use (~200ms)
2. **Large files**: CodeMirror handles files up to 10MB efficiently
3. **Multiple editors**: Reuse modules across instances (cached)

## Testing

```javascript
import { render, fireEvent } from '@testing-library/svelte';
import CodeEditor from './CodeEditor.svelte';

test('emits change event', async () => {
  const { component } = render(CodeEditor, {
    filename: 'test.tf',
    value: 'initial'
  });
  
  let changed = false;
  component.$on('change', (e) => {
    expect(e.detail).toBe('new content');
    changed = true;
  });
  
  // Simulate change...
  expect(changed).toBe(true);
});
```

## Troubleshooting

### Editor not loading
- Check browser console for errors
- Verify dependencies installed: `npm install codemirror codemirror-lang-terraform`

### Syntax highlighting not working
- Verify filename has `.tf` or `.tfvars` extension
- Check console for syntax loading errors

### Changes not saving
- Ensure you're handling the `change` event
- Remember to trigger reactivity for nested objects

### Fallback textarea appears
- Check error message above textarea
- Verify all dependencies installed correctly

## API Reference

### Methods

The component doesn't expose public methods. All interaction is through props and events.

### Lifecycle

1. **Mount**: Load modules → Initialize editor
2. **Update**: Respond to prop changes
3. **Destroy**: Clean up editor instance

### Internal State

- `editorView`: CodeMirror instance
- `isLoading`: Module loading state
- `useFallbackEditor`: Whether to use textarea fallback
- `loadError`: Error message (if any)

## Dependencies

```json
{
  "codemirror": "^6.0.0",
  "@codemirror/state": "^6.0.0",
  "@codemirror/language": "^6.0.0",
  "@codemirror/view": "^6.0.0",
  "@codemirror/commands": "^6.0.0",
  "codemirror-lang-terraform": "^1.0.0"
}
```

## Browser Support

- Chrome 90+
- Firefox 88+
- Safari 14+
- Edge 90+

## Related Files

- `CodeEditor.svelte` - Main component
- `theme.js` - Theme configuration
- `syntaxTestUtils.js` - Testing utilities
- `README.md` - Full documentation
- `INTEGRATION-GUIDE.md` - Integration guide

## Examples

See the `CodeEditor.demo.svelte` file for a complete working example.

## Support

For issues or questions:
1. Check the full [README](./README.md)
2. Review [Integration Guide](./INTEGRATION-GUIDE.md)
3. Check existing tests for examples
4. Review the [Design Document](../../.kiro/specs/terraform-syntax-highlighting/design.md)
