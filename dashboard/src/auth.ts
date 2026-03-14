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
        async jwt({ token, account }) {
            // Pin identity to the permanent GitHub Provider Account ID
            if (account) {
                token.id = account.providerAccountId;
            }
            return token;
        },
        async session({ session, token }) {
            if (token.id && session.user) {
                (session.user as any).id = token.id;
            }
            return session;
        },
    }
});
