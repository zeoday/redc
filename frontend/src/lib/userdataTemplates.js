import { ListUserdataTemplates } from '../../wailsjs/go/main/App.js';

export async function loadUserdataTemplates() {
  try {
    const result = await ListUserdataTemplates();
    return result || [];
  } catch (e) {
    console.error('Failed to load userdata templates:', e);
    return [];
  }
}

export function getTemplatesByType(templates, type) {
  if (!Array.isArray(templates)) return [];
  return templates.filter(t => t.type === type);
}

export function getTemplatesByCategory(templates, category) {
  if (!Array.isArray(templates)) return [];
  return templates.filter(t => t.category === category);
}

export function getAIScenarios(templates) {
  if (!Array.isArray(templates)) return [];
  return templates.filter(t => t.category === 'ai');
}
