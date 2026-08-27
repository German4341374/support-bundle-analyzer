<?php

declare(strict_types=1);

namespace SupportBundleAnalyzer\PhpWeb\Tests;

use PHPUnit\Framework\Attributes\CoversClass;
use PHPUnit\Framework\TestCase;
use SupportBundleAnalyzer\PhpWeb\Analyzer;

#[CoversClass(Analyzer::class)]
final class AnalyzerTest extends TestCase
{
    public function testGroupsFatalErrorsAndMemoryExhaustion(): void
    {
        $file = tempnam(sys_get_temp_dir(), 'sba-php-');
        self::assertIsString($file);
        file_put_contents($file, "PHP Fatal error: Uncaught Error\nAllowed memory size of 128 bytes exhausted\n");
        $findings = (new Analyzer())->analyze($file, 'php-error.log');
        unlink($file);
        self::assertCount(2, $findings);
        self::assertSame('PHP_FATAL_ERROR', $findings[0]->ruleId);
    }

    public function testDetectsRepeatedHttpServerErrors(): void
    {
        $file = tempnam(sys_get_temp_dir(), 'sba-web-');
        self::assertIsString($file);
        file_put_contents($file, str_repeat('127.0.0.1 - - "GET / HTTP/1.1" 503 12' . PHP_EOL, 5));
        $findings = (new Analyzer())->analyze($file, 'access.log');
        unlink($file);
        self::assertCount(1, $findings);
        self::assertSame('WEB_REPEATED_5XX', $findings[0]->ruleId);
    }

    public function testReturnsNoFindingForInformationalLog(): void
    {
        $file = tempnam(sys_get_temp_dir(), 'sba-clean-');
        self::assertIsString($file);
        file_put_contents($file, "application started\n");
        self::assertSame([], (new Analyzer())->analyze($file, 'app.log'));
        unlink($file);
    }
}
