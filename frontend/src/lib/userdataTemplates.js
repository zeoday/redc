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

export function getVulhubScenarios(templates) {
  if (!Array.isArray(templates)) return [];
  return templates.filter(t => t.category === 'vulhub');
}

export function getC2Scenarios(templates) {
  if (!Array.isArray(templates)) return [];
  return templates.filter(t => t.category === 'c2');
}

export function getGroupedTemplates(templates) {
  if (!Array.isArray(templates)) return {};
  const groups = {};
  for (const template of templates) {
    const cat = template.category || 'other';
    if (!groups[cat]) {
      groups[cat] = [];
    }
    groups[cat].push(template);
  }
  return groups;
}

export const userdataCategoryNames = {
  basic: '基础环境',
  ai: 'AI 应用',
  security: '安全工具',
  vulhub: '漏洞环境',
  c2: 'C2 场景',
  other: '其他'
};
