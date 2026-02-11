export interface QuickCommand {
  id: number;
  name: string;
  command: string;
  description?: string;
  groupId: number;
  sort: number;
  createdAt: string;
  updatedAt: string;
}

export interface QuickCommandCreate {
  name: string;
  command: string;
  description?: string;
  groupId?: number;
  sort?: number;
}

export interface QuickCommandUpdate {
  id: number;
  name?: string;
  command?: string;
  description?: string;
  groupId?: number;
  sort?: number;
}

export interface QuickCommandDelete {
  ids: number[];
}

export interface QuickCommandSearch {
  page: number;
  pageSize: number;
  keyword?: string;
  groupId?: number;
}

export interface QuickCommandPageResult {
  items: QuickCommand[];
  total: number;
}

export interface CommandTreeItem {
  value: string;
  label: string;
  children?: CommandTreeItem[];
}