import { useStoredTheme } from './hooks/useStoredTheme';
import {
  BrowserRouter as Router,
  Navigate,
  Route,
  Routes,
} from 'react-router-dom';
import { StartRoute } from './routes/Start';
import styled, { createGlobalStyle, ThemeProvider } from 'styled-components';
import { Notifications } from './components/Notifications';
import React from 'react';
import { DashboardRoute } from './routes/Dashboard';
import { DebugRoute } from './routes/Debug';
import { RouteSuspense } from './components/RouteSuspense';

const LoginRoute = React.lazy(() => import('./routes/Login'));
const GuildMembersRoute = React.lazy(
  () => import('./routes/Dashboard/Guilds/GuildMembers')
);
const MemberRoute = React.lazy(
  () => import('./routes/Dashboard/Guilds/Member')
);

const GlobalStyle = createGlobalStyle`
  body {
    background-color: ${(p) => p.theme.background};
    color: ${(p) => p.theme.text};
  }

  * {
    box-sizing: border-box;
  }
`;

const AppContainer = styled.div`
  width: 100vw;
  height: 100vh;
`;

export const App: React.FC = () => {
  const { theme } = useStoredTheme();

  return (
    <ThemeProvider theme={theme}>
      <AppContainer>
        <Router>
          <Routes>
            <Route path="start" element={<StartRoute />} />
            <Route path="login" element={<LoginRoute />} />
            <Route path="db" element={<DashboardRoute />}>
              <Route
                path="guilds/:guildid/members"
                element={
                  <RouteSuspense>
                    <GuildMembersRoute />
                  </RouteSuspense>
                }
              />
              <Route
                path="guilds/:guildid/members/:memberid"
                element={
                  <RouteSuspense>
                    <MemberRoute />
                  </RouteSuspense>
                }
              />
              {import.meta.env.DEV && (
                <Route path="debug" element={<DebugRoute />} />
              )}
            </Route>
            <Route path="*" element={<Navigate to="db" />} />
          </Routes>
        </Router>
        <Notifications />
      </AppContainer>
      <GlobalStyle />
    </ThemeProvider>
  );
};

export default App;
