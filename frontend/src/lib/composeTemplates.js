import { ListComposeTemplates } from '../../wailsjs/go/main/App.js';

export async function loadComposeTemplates() {
  try {
    const result = await ListComposeTemplates();
    return result || [];
  } catch (e) {
    console.error('Failed to load compose templates:', e);
    return [];
  }
}

export function getComposeTemplatesByCategory(templates, category) {
  if (!Array.isArray(templates)) return [];
  return templates.filter(t => t.category === category);
}
