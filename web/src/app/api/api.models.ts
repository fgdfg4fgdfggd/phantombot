/** @format */

export interface ListReponse<T> {
  n: number;
  data: T[];
}

export interface User {
  id: string;
  username: string;
  avatar: string;
  locale: string;
  discriminator: string;
  verified: boolean;
  bot: boolean;
  avatar_url: string;
  created_at?: string;
}

export interface Role {
  id: string;
  name: string;
  managed: boolean;
  mentionable: boolean;
  hoist: boolean;
  color: number;
  position: number;
  permission: number;
}

export interface Member {
  user: User;
  guild_id: string;
  joined_at: string;
  nick: string;
  avatar_url?: string;
  roles: string[];
  created_at?: string;
}

export interface Channel {
  id: string;
  guild_id: string;
  name: string;
  topic: string;
  type: number;
  nsfw: boolean;
  position: number;
  user_limit: number;
  parent_id: string;
}

export interface Guild {
  id: string;
  name: string;
  icon: string;
  icon_url: string;
  region: string;
  owner_id: string;
  joined_at: string;
  member_count: number;

  self_member?: Member;

  roles?: Role[];
  members?: Member[];
  channels?: Channel[];
}

export interface PermissionResponse {
  permissions: number;
}

export interface Report {
  id: string;
  type: number;
  type_name?: string;
  guild_id: string;
  executor_id: string;
  victim_id: string;
  message: string;
  attachment_url: string;
  created?: string;
}
