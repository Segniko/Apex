import NextAuth from "next-auth";
import GitHub from "next-auth/providers/github";

export const { handlers, auth, signIn, signOut } = NextAuth({
    providers: [
        GitHub,
    ],
    pages: {
        signIn: "/auth/login",
    },
    session: {
        strategy: "jwt",
        maxAge: 1 * 24 * 60 * 60, // 1 day
    },
    trustHost: true,
    callbacks: {
        session({ session, token }) {
            if (token.sub && session.user) {
                (session.user as any).id = token.sub;
            }
            return session;
        },
    }
});
