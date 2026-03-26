import type { LucideIcon } from "lucide-react";
import { Files, LayoutDashboard, MessageSquareQuote, Settings } from "lucide-react";

export type NavigationItem = {
  href: string;
  label: string;
  description: string;
  icon: LucideIcon;
};

export const navigationItems: NavigationItem[] = [
  {
    href: "/",
    label: "Дашборд",
    description: "Ключевые метрики и состояние платформы.",
    icon: LayoutDashboard,
  },
  {
    href: "/documents",
    label: "Документы",
    description: "Загрузка и мониторинг задач индексации.",
    icon: Files,
  },
  {
    href: "/chat",
    label: "RAG-чат",
    description: "Семантические ответы по вашим данным.",
    icon: MessageSquareQuote,
  },
  {
    href: "/settings",
    label: "Настройки",
    description: "Параметры рабочего пространства и команды.",
    icon: Settings,
  },
];
