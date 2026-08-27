<?php

declare(strict_types=1);

namespace SupportBundleAnalyzer\PhpWeb;

final class Analyzer
{
    /** @var array<string, array{severity: string, title: string, pattern: string}> */
    private const RULES = [
        'PHP_FATAL_ERROR' => ['severity' => 'high', 'title' => 'Fatal PHP errors observed', 'pattern' => '/PHP Fatal error|Uncaught (?:Error|Exception)/i'],
        'PHP_MEMORY_LIMIT' => ['severity' => 'high', 'title' => 'PHP memory limit exhausted', 'pattern' => '/Allowed memory size .* exhausted/i'],
        'PHP_EXECUTION_TIMEOUT' => ['severity' => 'medium', 'title' => 'PHP execution time exceeded', 'pattern' => '/Maximum execution time .* exceeded/i'],
        'PHP_FPM_CAPACITY' => ['severity' => 'high', 'title' => 'PHP-FPM worker capacity exhausted', 'pattern' => '/server reached pm\.max_children|pool .* seems busy/i'],
        'WEB_UPSTREAM_TIMEOUT' => ['severity' => 'high', 'title' => 'Web upstream timeout observed', 'pattern' => '/upstream timed out|gateway timeout/i'],
        'WEB_PERMISSION_DENIED' => ['severity' => 'medium', 'title' => 'Web process permission denial observed', 'pattern' => '/permission denied/i'],
    ];

    /** @return list<Finding> */
    public function analyze(string $filename, string $artifactPath): array
    {
        $handle = fopen($filename, 'rb');
        if (false === $handle) {
            throw new \RuntimeException('Artifact cannot be opened.');
        }
        /** @var array<string, array{count: int, firstLine: int, excerpt: string}> $matches */
        $matches = [];
        $lineNumber = 0;
        $http5xx = 0;
        while (false !== ($line = fgets($handle))) {
            ++$lineNumber;
            if ($lineNumber > 2_000_000) {
                break;
            }
            foreach (self::RULES as $ruleId => $rule) {
                if (1 !== preg_match($rule['pattern'], $line)) {
                    continue;
                }
                $match = $matches[$ruleId] ?? ['count' => 0, 'firstLine' => $lineNumber, 'excerpt' => self::excerpt($line)];
                ++$match['count'];
                $matches[$ruleId] = $match;
            }
            if (1 === preg_match('/"\s+5\d{2}\s+/', $line)) {
                ++$http5xx;
            }
        }
        fclose($handle);
        if ($http5xx >= 5) {
            $matches['WEB_REPEATED_5XX'] = ['count' => $http5xx, 'firstLine' => 1, 'excerpt' => 'Repeated HTTP 5xx status records'];
        }
        $findings = [];
        ksort($matches);
        foreach ($matches as $ruleId => $match) {
            $definition = self::RULES[$ruleId] ?? ['severity' => 'medium', 'title' => 'Repeated HTTP 5xx responses'];
            $findings[] = new Finding(
                $ruleId,
                $definition['severity'],
                $definition['title'],
                sprintf('Observed %d matching record(s). The surrounding application and platform evidence needs review.', $match['count']),
                'strong',
                [['artifact' => $artifactPath, 'lineStart' => $match['firstLine'], 'excerpt' => $match['excerpt']]],
            );
        }

        return $findings;
    }

    private static function excerpt(string $line): string
    {
        $clean = trim($line);

        return substr($clean, 0, 300);
    }
}
