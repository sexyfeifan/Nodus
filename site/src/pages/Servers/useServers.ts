import { useState, useEffect, useCallback, useRef } from "react";
import pb from "../../lib/pocketbase";
import { apiGet, apiPost } from "../../lib/api";
import { toast } from "sonner";
import { useTranslation } from "react-i18next";

export interface Server {
  id: string;
  serverName: string;
  user: string;
  serverAddr: string;
  serverPort: number;
  description: string;
  bootStatus: string;
  autoConnection: boolean;
  created: string;
  updated: string;
  sendRate: number;
  recvRate: number;
  // Network status
  networkStatus?: {
    latency: number;
    reachable: boolean;
    lastCheckTime: string;
    error?: string;
  };
  geoLocation?: {
    country: string;
    countryCode: string;
    region: string;
    city: string;
    isp: string;
    lat: number;
    lon: number;
  };
  // Extended fields
  log?: {
    level: "trace" | "debug" | "info" | "warn" | "error";
    maxDays: number;
    to?: string;
  };
  auth?: {
    method: "none" | "token" | "oidc";
    token: string;
    oidcClientId: string;
    oidcClientSecret: string;
    oidcAudience: string;
    oidcTokenEndpoint: string;
  };
  transport?: {
    protocol: "tcp" | "kcp" | "quic" | "websocket" | "wss";
    tls: {
      enable: boolean;
      disableCustomTLSFirstByte: boolean;
      certFile: string;
      keyFile: string;
      trustedCaFile: string;
      serverName: string;
    };
    proxyURL: string;
  };
  metadatas?: Record<string, string>;
}

export function useServers() {
  const { t } = useTranslation();
  const [servers, setServers] = useState<Server[]>([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const initializedRef = useRef(false);
  const mountedRef = useRef(true);
  const requestIdRef = useRef(0);

  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const PER_PAGE = 10;

  const [search, setSearch] = useState("");

  useEffect(() => {
    mountedRef.current = true;
    return () => {
      mountedRef.current = false;
    };
  }, []);

  const fetchServers = useCallback(
    async (isRefresh = false) => {
      const requestId = ++requestIdRef.current;
      try {
        if (!initializedRef.current) {
          setLoading(true);
        } else if (isRefresh) {
          setRefreshing(true);
        }

        const params = new URLSearchParams({
          page: String(page),
          perPage: String(PER_PAGE),
        });
        if (search) params.set("search", search);

        const res = await apiGet(`/api/servers?${params}`);
        if (!res.ok) throw new Error("Failed to fetch servers");
        const data = await res.json();

        if (!mountedRef.current || requestId !== requestIdRef.current) return;
        setServers(data.items);
        setTotalPages(data.totalPages);
        initializedRef.current = true;
      } catch (err) {
        if ((err as any)?.isAbort) return;
        if (!mountedRef.current || requestId !== requestIdRef.current) return;
        toast.error(err instanceof Error ? err.message : t("server.fetchFailed"));
      } finally {
        if (!mountedRef.current || requestId !== requestIdRef.current) return;
        setLoading(false);
        setRefreshing(false);
      }
    },
    [page, search, t]
  );

  useEffect(() => {
    setPage(1);
    if (initializedRef.current) {
      setLoading(true);
    }
  }, [search]);

  useEffect(() => {
    const timer = setTimeout(() => {
      fetchServers();
    }, 300); // Simple debounce
    return () => clearTimeout(timer);
  }, [fetchServers]);

  // Auto-refresh server status every 5 seconds
  useEffect(() => {
    const interval = setInterval(() => {
      fetchServers(true); // Pass true to indicate this is a refresh
    }, 5000);
    return () => clearInterval(interval);
  }, [fetchServers]);

  const deleteServer = async (id: string) => {
    try {
      await pb.collection("fh_servers").delete(id);
      await fetchServers();
      toast.success(t("server.deleteSuccess"));
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t("server.deleteFailed"));
    }
  };

  const launchServer = async (id: string) => {
    try {
      const response = await apiPost("/api/frpc/launch", { id });
      if (!response.ok) {
        toast.error(t("server.launchFailed"));
        return;
      }
      await fetchServers();
      toast.success(t("server.launchSuccess"));
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t("server.launchFailed"));
    }
  };

  const terminateServer = async (id: string) => {
    try {
      const response = await apiPost("/api/frpc/terminate", { id });
      if (!response.ok) {
        toast.error(t("server.terminateFailed"));
        return;
      }
      await fetchServers();
      toast.success(t("server.terminateSuccess"));
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t("server.terminateFailed"));
    }
  };

  return {
    servers,
    loading,
    refreshing,
    deleteServer,
    page,
    setPage,
    totalPages,
    launchServer,
    terminateServer,
    search,
    setSearch,
    refreshServers: () => fetchServers(true),
  };
}
