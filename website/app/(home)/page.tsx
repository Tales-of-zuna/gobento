import Link from 'next/link';
import { MascotHero } from '@/components/mascot/MascotHero';
import { Button } from '@/components/ui/button';

export default function HomePage() {
  return (
    <div className="flex flex-1 flex-col items-center justify-center gap-4 px-4 py-8 text-center">
      <MascotHero className="max-w-[min(720px,92vw)]" />
      <h1 className="text-2xl font-bold">gobento</h1>
      <p>Compose a ready-to-build Go project from the stack you choose.</p>
      <Button nativeButton={false} render={<Link href="/docs/" />}>
        Read the docs
      </Button>
    </div>
  );
}
