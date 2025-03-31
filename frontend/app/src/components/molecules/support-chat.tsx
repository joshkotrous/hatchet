import { User } from '@/lib/api';
import { useTenant } from '@/lib/atoms';
import useApiMeta from '@/pages/auth/hooks/use-api-meta';
import React, { PropsWithChildren, useEffect, useMemo } from 'react';

interface SupportChatProps {
  user: User;
}

const SupportChat: React.FC<PropsWithChildren & SupportChatProps> = ({
  user,
  children,
}) => {
  const meta = useApiMeta();

  const { tenant } = useTenant();

  const APP_ID = useMemo(() => {
    if (!meta.data?.pylonAppId) {
      return null;
    }

    // Validate that pylonAppId only contains expected characters
    const pylonAppId = meta.data.pylonAppId;
    if (typeof pylonAppId !== 'string' || !/^[a-zA-Z0-9_-]+$/.test(pylonAppId)) {
      console.error('Invalid pylonAppId format');
      return null;
    }

    return pylonAppId;
  }, [meta]);

  useEffect(() => {
    if (!APP_ID) {
      return;
    }

    // Reimplement the Pylon initialization without using innerHTML
    const e = window;
    const t = document;
    
    // Create the Pylon function
    const n = function() {
      (n as any).e(arguments);
    } as any;
    n.q = [];
    n.e = function(e: any) {
      n.q.push(e);
    };
    e.Pylon = n;
    
    // Create the function that loads the script
    const r = function() {
      const e = t.createElement("script");
      e.setAttribute("type", "text/javascript");
      e.setAttribute("async", "true");
      e.setAttribute("src", `https://widget.usepylon.com/widget/${APP_ID}`);
      const n = t.getElementsByTagName("script")[0];
      if (n && n.parentNode) {
        n.parentNode.insertBefore(e, n);
      }
    };
    
    // Execute based on document readiness
    if (t.readyState === "complete") {
      r();
    } else if (e.addEventListener) {
      e.addEventListener("load", r, false);
    }
  }, [APP_ID]);

  useEffect(() => {
    if (!APP_ID || !user) {
      return;
    }

    (window as any).pylon = {
      chat_settings: {
        app_id: APP_ID,
        email: user.email,
        name: user.name,
        email_hash: user.emailHash,
      },
    };

    (window as any).Pylon('setNewIssueCustomFields', {
      user_id: user.metadata.id,
      tenant_name: tenant?.name,
      tenant_slug: tenant?.slug,
      tenant_id: tenant?.metadata?.id,
    });
  }, [user, APP_ID, tenant]);

  return children;
};

export default SupportChat;