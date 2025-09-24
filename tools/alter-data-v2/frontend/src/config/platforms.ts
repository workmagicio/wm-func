export interface Platform {
  id: string;
  name: string;
  displayName: string;
  icon: string;
  color: string;
  enabled: boolean;
}

export const PLATFORMS: Platform[] = [
  {
    id: 'googleAds',
    name: 'Google Ads',
    displayName: 'Google Ads',
    icon: '🔍',
    color: '#4285f4',
    enabled: true
  },
  {
    id: 'facebookMarketing',
    name: 'Facebook Marketing',
    displayName: 'Meta Ads',
    icon: '📘',
    color: '#1877f2',
    enabled: true
  },
  {
    id: 'tiktokMarketing',
    name: 'TikTok Marketing',
    displayName: 'TikTok Ads',
    icon: '🎵',
    color: '#ff0050',
    enabled: true
  },
  {
    id: 'snapchatMarketing',
    name: 'Snapchat Marketing',
    displayName: 'Snapchat Ads',
    icon: '👻',
    color: '#fffc00',
    enabled: true
  },
  {
    id: 'pinterest',
    name: 'Pinterest',
    displayName: 'Pinterest Ads',
    icon: '🎨',
    color: '#e60023',
    enabled: true
  },
  {
    id: 'applovin',
    name: 'AppLovin',
    displayName: 'AppLovin Ads',
    icon: '🚀',
    color: '#000000',
    enabled: true
  },
  {
    id: 'bingAds',
    name: 'BingAds',
    displayName: 'BingAds Ads',
    icon: '🔍',
    color: '#0078d7',
    enabled: true
  },
  {
    id: 'amazonVendorPartner',
    name: 'Amazon Vendor Partner',
    displayName: 'Amazon Vendor Partner',
    icon: '📦',
    color: '#ff9900',
    enabled: true
  },
  {
    id: 'fairing',
    name: 'Fairing',
    displayName: 'Fairing Survey',
    icon: '📊',
    color: '#6366f1',
    enabled: true
  },
  {
    id: 'amazonAds',
    name: 'Amazon Ads',
    displayName: 'Amazon Ads',
    icon: '🛒',
    color: '#ff9900',
    enabled: true
  },
  {
    id: 'knocommerce',
    name: 'Knocommerce',
    displayName: 'Knocommerce',
    icon: '📈',
    color: '#8b5cf6',
    enabled: true
  },
  {
    id: 'applovinLog',
    name: 'AppLovin Log',
    displayName: 'AppLovin Log',
    icon: '📝',
    color: '#2563eb',
    enabled: true
  }
];

// 获取启用的平台
export const getEnabledPlatforms = (): Platform[] => {
  return PLATFORMS.filter(platform => platform.enabled);
};

// 根据ID获取平台信息
export const getPlatformById = (id: string): Platform | undefined => {
  return PLATFORMS.find(platform => platform.id === id);
};

// 获取平台显示名称
export const getPlatformDisplayName = (id: string): string => {
  const platform = getPlatformById(id);
  return platform ? platform.displayName : id;
};
