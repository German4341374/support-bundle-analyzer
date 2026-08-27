<?php

declare(strict_types=1);

namespace SupportBundleAnalyzer\PhpWeb;

final readonly class Finding implements \JsonSerializable
{
    /** @param list<array{artifact: string, lineStart: int, excerpt: string}> $evidence */
    public function __construct(
        public string $ruleId,
        public string $severity,
        public string $title,
        public string $summary,
        public string $confidence,
        public array $evidence,
    ) {}

    /** @return array<string, mixed> */
    public function jsonSerialize(): array
    {
        return get_object_vars($this);
    }
}
