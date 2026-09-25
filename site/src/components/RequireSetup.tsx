import { Navigate, useLocation } from "react-router-dom";
import { useEffect, useState } from "react";
import { Button, Flex, Text } from "@radix-ui/themes";
import { Icon } from "@iconify/react";
import { useTranslation } from "react-i18next";
import { Loading } from "./Loading";
import { useSystemStore } from "../stores/systemStore";

interface RequireSetupProps {
  children: React.ReactNode;
}

/**
 * Route guard component
 * Check if system has been initialized by checking the settings table
 * If not initialized, redirect to setup page
 */
export function RequireSetup({ children }: RequireSetupProps) {
  const location = useLocation();
  const { t } = useTranslation();
  const { checkInitialized } = useSystemStore();
  const [isInitialized, setIsInitialized] = useState<boolean | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [checkFailed, setCheckFailed] = useState(false);

  useEffect(() => {
    checkInitialization();
  }, []);

  const checkInitialization = async () => {
    setIsLoading(true);
    setCheckFailed(false);
    try {
      const initialized = await checkInitialized();
      setIsInitialized(initialized);
    } catch (error) {
      console.error("Failed to check initialization status:", error);
      setCheckFailed(true);
    } finally {
      setIsLoading(false);
    }
  };

  if (isLoading) {
    return <Loading fullscreen />;
  }

  if (checkFailed) {
    return (
      <Flex
        direction="column"
        align="center"
        justify="center"
        gap="4"
        style={{ minHeight: "100vh", backgroundColor: "var(--gray-2)" }}
      >
        <Icon icon="lucide:alert-triangle" color="var(--red-9)" width="48" height="48" />
        <Text size="3" weight="medium">
          {t("error.somethingWrong")}
        </Text>
        <Button onClick={checkInitialization}>{t("error.tryAgain")}</Button>
      </Flex>
    );
  }

  if (!isInitialized) {
    // Not initialized, redirect to setup page
    return <Navigate to="/setup" state={{ from: location }} replace />;
  }

  return <>{children}</>;
}
